import {useCallback, useEffect, useRef, useState} from "react";
import {getCourse} from "../api";
import {requirementKey, requirementOptionLabel} from "../prerequisites";
import type {CourseSummary, DependencyNode, DependencyTree} from "../util";
import {courseName} from "../util";

type LoadEntry =
    | {status: "loading"}
    | {status: "loaded"; tree: DependencyTree}
    | {status: "error"; message: string};

interface RequirementProps {
    node: DependencyNode;
    loads: Map<string, LoadEntry>;
    loadCourse: (course: CourseSummary) => void;
}

export default function PrerequisitePaths({initialTree}: {initialTree: DependencyTree}) {
    const loads = useRef(new Map<string, LoadEntry>());
    const generation = useRef(0);
    const [, render] = useState(0);

    useEffect(() => {
        generation.current += 1;
        loads.current.clear();
        render((value) => value + 1);
    }, [initialTree]);

    const loadCourse = useCallback((course: CourseSummary) => {
        const key = courseName(course).toUpperCase();
        const existing = loads.current.get(key);
        if (existing && existing.status !== "error") return;

        const requestGeneration = generation.current;
        loads.current.set(key, {status: "loading"});
        render((value) => value + 1);
        void getCourse(course.course_category, course.course_code, 1)
            .then((detail) => {
                if (generation.current !== requestGeneration) return;
                loads.current.set(key, {status: "loaded", tree: detail.tree ?? emptyTree(detail.course)});
                render((value) => value + 1);
            })
            .catch((reason: unknown) => {
                if (generation.current !== requestGeneration) return;
                const message = reason instanceof Error ? reason.message : "Unable to load prerequisites.";
                loads.current.set(key, {status: "error", message});
                render((value) => value + 1);
            });
    }, []);

    const children = initialTree.root.children ?? [];
    if (children.length === 0) {
        return <div className="empty-state">No course prerequisites listed.</div>;
    }

    return (
        <div className="prerequisite-paths">
            <RequirementList nodes={children} loads={loads.current} loadCourse={loadCourse}/>
        </div>
    );
}

function RequirementList({nodes, loads, loadCourse}: {nodes: DependencyNode[]} & Omit<RequirementProps, "node">) {
    return (
        <div className="requirement-list">
            {nodes.map((node, index) => (
                <Requirement key={requirementKey(node, index)} node={node} loads={loads} loadCourse={loadCourse}/>
            ))}
        </div>
    );
}

function Requirement({node, loads, loadCourse}: RequirementProps) {
    if (node.type === "course") {
        return <CoursePath node={node} loads={loads} loadCourse={loadCourse}/>;
    }
    if (node.type === "and") {
        return <AndGroup children={node.children ?? []} loads={loads} loadCourse={loadCourse}/>;
    }
    return <OrGroup children={node.children ?? []} loads={loads} loadCourse={loadCourse}/>;
}

function AndGroup({children, loads, loadCourse}: {children: DependencyNode[]} & Omit<RequirementProps, "node">) {
    const alignsOrChoices = children.some((child) => child.type === "or");

    return (
        <section className="logic-group and-group" aria-label="All required courses">
            <p className="logic-label"><span>AND</span> All required</p>
            <div className={`and-columns${alignsOrChoices ? " has-or-choice" : ""}`}>
                {children.map((child, index) => (
                    <div className={`and-column${child.type === "or" ? " is-or-column" : ""}`} key={requirementKey(child, index)}>
                        {alignsOrChoices && child.type !== "or" && <div className="and-choice-spacer" aria-hidden="true"/>}
                        <Requirement node={child} loads={loads} loadCourse={loadCourse}/>
                    </div>
                ))}
            </div>
        </section>
    );
}

function OrGroup({children, loads, loadCourse}: {children: DependencyNode[]} & Omit<RequirementProps, "node">) {
    const [selected, setSelected] = useState("");
    const selectedIndex = selected === "" ? -1 : Number(selected);

    return (
        <section className="logic-group or-group" aria-label="Choose one course">
            <div className="or-choice-control">
                <label>
                    <span className="logic-label"><span>OR</span> Choose one course</span>
                    <select value={selected} onChange={(event) => setSelected(event.target.value)}>
                        <option value="">Choose a course</option>
                        {children.map((child, index) => (
                            <option value={index} key={requirementKey(child, index)}>
                                {requirementOptionLabel(child, index)}
                            </option>
                        ))}
                    </select>
                </label>
            </div>
            {selectedIndex >= 0 && children[selectedIndex] && (
                <div className="or-selection" key={requirementKey(children[selectedIndex], selectedIndex)}>
                    <Requirement node={children[selectedIndex]} loads={loads} loadCourse={loadCourse}/>
                </div>
            )}
            {selectedIndex < 0 && (
                <div className="or-selection or-selection-placeholder" aria-hidden="true">
                    <div className="path-course-card path-course-placeholder">
                        <span className="path-course-main">
                            <span className="path-course-code placeholder-code">&nbsp;</span>
                            <strong className="placeholder-title">&nbsp;</strong>
                        </span>
                        <span className="path-course-action placeholder-action"/>
                    </div>
                </div>
            )}
        </section>
    );
}

function CoursePath({node, loads, loadCourse}: RequirementProps) {
    const [open, setOpen] = useState(false);
    const course = node.course;
    if (!course) return <div className="path-message path-error">Course information is unavailable.</div>;
    const selectedCourse = course;

    const key = courseName(selectedCourse).toUpperCase();
    const entry = loads.get(key);
    const embeddedChildren = node.children ?? [];
    const canExpand = !node.missing && !node.cycle && (node.truncated === true || embeddedChildren.length > 0);

    function toggle() {
        const nextOpen = !open;
        setOpen(nextOpen);
        if (nextOpen && embeddedChildren.length === 0 && (!entry || entry.status === "error")) loadCourse(selectedCourse);
    }

    const courseContent = (
        <span className="path-course-main">
            <span className="path-course-code">{courseName(selectedCourse)}</span>
            <strong>{selectedCourse.course_title || "Course title unavailable"}</strong>
        </span>
    );

    if (!canExpand) {
        const staticLabel = node.missing ? "Not in catalog" : node.cycle ? "Already shown above" : "No prerequisites";
        return (
            <div className="course-path">
                <div className="path-course-card is-static" aria-label={`${courseName(selectedCourse)} ${selectedCourse.course_title || "Course"} ${staticLabel}`}>
                    {courseContent}
                    {(node.missing || node.cycle) && (
                        <span className="path-course-action" aria-hidden="true">{node.missing ? "!" : "↩"}</span>
                    )}
                </div>
            </div>
        );
    }

    return (
        <div className="course-path">
            <button
                className="path-course-card"
                type="button"
                onClick={toggle}
                aria-expanded={open}
                aria-label={`${courseName(selectedCourse)} ${selectedCourse.course_title || "Course"} ${open ? "Collapse prerequisites" : "Expand prerequisites"}`}
            >
                {courseContent}
                <span className={`path-course-action${open ? " is-open" : ""}`} aria-hidden="true">
                    {open ? "−" : "+"}
                </span>
            </button>

            {open && (
                <div className="path-children">
                    {embeddedChildren.length > 0 ? (
                        <RequirementList nodes={embeddedChildren} loads={loads} loadCourse={loadCourse}/>
                    ) : entry?.status === "loading" ? (
                        <div className="path-message">Loading…</div>
                    ) : entry?.status === "error" ? (
                        <div className="path-message path-error">Could not load this course. Close and open it to retry. {entry.message}</div>
                    ) : entry?.status === "loaded" && (entry.tree.root.children ?? []).length > 0 ? (
                        <RequirementList nodes={entry.tree.root.children ?? []} loads={loads} loadCourse={loadCourse}/>
                    ) : entry?.status === "loaded" ? (
                        <div className="path-message">No more course prerequisites.</div>
                    ) : null}
                </div>
            )}
        </div>
    );
}

function emptyTree(course: CourseSummary): DependencyTree {
    return {max_depth: 1, root: {type: "course", course, children: []}};
}
