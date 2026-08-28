import type {DependencyNode} from "./util";
import {courseName} from "./util";

export function requirementOptionLabel(node: DependencyNode, index: number): string {
    if (node.type === "course" && node.course) {
        const title = node.course.course_title ? ` — ${node.course.course_title}` : "";
        return `${courseName(node.course)}${title}`;
    }

    const courses = collectCourseNames(node);
    if (courses.length === 0) return `Option ${index + 1}`;
    if (node.type === "and") return courses.join(" + ");
    return courses.join(" or ");
}

export function requirementKey(node: DependencyNode, index: number): string {
    if (node.type === "course" && node.course) return `${courseName(node.course)}-${index}`;
    return `${node.type}-${index}-${collectCourseNames(node).join("-")}`;
}

function collectCourseNames(node: DependencyNode): string[] {
    if (node.type === "course") return node.course ? [courseName(node.course)] : [];
    return (node.children ?? []).flatMap(collectCourseNames);
}
