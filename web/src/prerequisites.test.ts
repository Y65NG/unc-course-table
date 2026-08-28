import {describe, expect, it} from "vitest";
import {requirementKey, requirementOptionLabel} from "./prerequisites";
import type {DependencyNode} from "./util";

const course = (name: string, title = ""): DependencyNode => {
    const [course_category, course_code] = name.split(" ");
    return {
        type: "course",
        course: {course_category, course_code, course_title: title, course_credit: "3"},
    };
};

describe("prerequisite presentation helpers", () => {
    it("labels a course option with its code and title", () => {
        expect(requirementOptionLabel(course("MATH 231", "Calculus I"), 0)).toBe("MATH 231 — Calculus I");
    });

    it("summarizes a grouped OR option without exposing implementation terms", () => {
        const option: DependencyNode = {type: "and", children: [course("COMP 210"), course("COMP 211")]};
        expect(requirementOptionLabel(option, 0)).toBe("COMP 210 + COMP 211");
    });

    it("creates stable sibling keys for repeated course references", () => {
        expect(requirementKey(course("COMP 110"), 0)).not.toBe(requirementKey(course("COMP 110"), 1));
    });
});
