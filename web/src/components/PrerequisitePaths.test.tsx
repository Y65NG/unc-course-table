// @vitest-environment jsdom

import {cleanup, fireEvent, render, screen, waitFor} from "@testing-library/react";
import {afterEach, beforeEach, describe, expect, it, vi} from "vitest";
import {getCourse} from "../api";
import type {Course, CourseDetailResponse, CourseSummary, DependencyNode, DependencyTree} from "../util";
import PrerequisitePaths from "./PrerequisitePaths";

vi.mock("../api", () => ({getCourse: vi.fn()}));

const mockedGetCourse = vi.mocked(getCourse);

function summary(name: string, title: string): CourseSummary {
    const [course_category, course_code] = name.split(" ");
    return {course_category, course_code, course_title: title, course_credit: "3"};
}

function courseNode(course: CourseSummary, children: DependencyNode[] = [], truncated = false): DependencyNode {
    return {type: "course", course, children, truncated};
}

function detail(course: CourseSummary, children: DependencyNode[]): CourseDetailResponse {
    const fullCourse: Course = {...course, course_description: "", course_prerequisites: ""};
    return {course: fullCourse, tree: {max_depth: 1, root: courseNode(course, children)}};
}

describe("PrerequisitePaths", () => {
    beforeEach(() => mockedGetCourse.mockReset());
    afterEach(cleanup);

    it("resets the expanded state when an OR selection changes", async () => {
        const math130 = summary("MATH 130", "Precalculus Mathematics");
        const math152 = summary("MATH 152", "Calculus for Business");
        const math110 = summary("MATH 110", "Algebra");
        const root = summary("COMP 110", "Introduction to Programming");
        const tree: DependencyTree = {
            max_depth: 1,
            root: courseNode(root, [{type: "or", children: [courseNode(math130, [], true), courseNode(math152, [], true)]}]),
        };
        mockedGetCourse.mockResolvedValue(detail(math130, [courseNode(math110)]));

        render(<PrerequisitePaths initialTree={tree}/>);
        const select = screen.getByRole("combobox", {name: /choose one/i});
        fireEvent.change(select, {target: {value: "0"}});
        fireEvent.click(screen.getByRole("button", {name: /MATH 130/i}));
        await waitFor(() => expect(screen.getByText("MATH 110")).toBeTruthy());

        fireEvent.change(select, {target: {value: "1"}});

        expect(screen.getByRole("button", {name: /MATH 152/i}).getAttribute("aria-expanded")).toBe("false");
    });

    it("renders a course without prerequisites as a static row", () => {
        const root = summary("COMP 210", "Data Structures");
        const leaf = summary("MATH 110", "Algebra");
        const tree: DependencyTree = {max_depth: 1, root: courseNode(root, [courseNode(leaf)])};

        const {container} = render(<PrerequisitePaths initialTree={tree}/>);

        expect(screen.queryByRole("button", {name: /MATH 110/i})).toBeNull();
        expect(screen.getByText("MATH 110")).toBeTruthy();
        expect(container.querySelector(".path-course-action")).toBeNull();
    });

    it("aligns a selected OR course with its sibling AND course", () => {
        const root = summary("COMP 410", "Data Structures");
        const math231 = summary("MATH 231", "Calculus of Functions of One Variable I");
        const math241 = summary("MATH 241", "Calculus of Functions of Several Variables");
        const comp401 = summary("COMP 401", "Foundation of Programming");
        const tree: DependencyTree = {
            max_depth: 1,
            root: courseNode(root, [{
                type: "and",
                children: [
                    {type: "or", children: [courseNode(math231), courseNode(math241)]},
                    courseNode(comp401, [], true),
                ],
            }]),
        };

        const {container} = render(<PrerequisitePaths initialTree={tree}/>);

        const placeholder = container.querySelector(".is-or-column .path-course-placeholder");
        expect(placeholder).toBeTruthy();
        expect(placeholder?.classList.contains("path-course-card")).toBe(true);
        expect(placeholder?.querySelector(".placeholder-code")).toBeTruthy();
        expect(placeholder?.querySelector(".placeholder-title")).toBeTruthy();
        fireEvent.change(screen.getByRole("combobox", {name: /choose one/i}), {target: {value: "0"}});

        const columns = container.querySelectorAll(".and-columns.has-or-choice > .and-column");
        expect(columns).toHaveLength(2);
        expect(columns[0].classList.contains("is-or-column")).toBe(true);
        expect(columns[0].querySelector(".path-course-placeholder")).toBeNull();
        expect(columns[0].querySelector(":scope > .or-group > .or-selection .path-course-card")).toBeTruthy();
        expect(columns[1].querySelector(":scope > .and-choice-spacer")).toBeTruthy();
        expect(columns[1].querySelector(":scope > .course-path > .path-course-card")).toBeTruthy();
    });
});
