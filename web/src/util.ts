export type ParseStatus = "complete" | "partial" | "failed";
export type RequirementKind =
    | "prerequisite"
    | "corequisite"
    | "pre_or_corequisite"
    | "unspecified";

export interface CourseSummary {
    course_category: string;
    course_code: string;
    course_title: string;
    course_credit: string;
}

export interface Course extends CourseSummary {
    course_description: string;
    course_prerequisites: string;
}

export interface ParseDiagnostic {
    code: string;
    message: string;
    fragment?: string;
}

export interface ParseResult {
    status: ParseStatus;
    kind: RequirementKind;
    source: string;
    diagnostics: ParseDiagnostic[];
}

export interface DependencyNode {
    type: "course" | "and" | "or";
    course?: CourseSummary;
    children?: DependencyNode[];
    parse_status?: ParseStatus;
    kind?: RequirementKind;
    diagnostics?: ParseDiagnostic[];
    cycle?: boolean;
    truncated?: boolean;
    missing?: boolean;
}

export interface DependencyTree {
    root: DependencyNode;
    max_depth: number;
}

export interface SearchResponse {
    query: string;
    results: CourseSummary[];
    page: number;
    page_size: number;
    total: number;
    total_pages: number;
}

export interface CourseDetailResponse {
    course: Course;
    parse?: ParseResult;
    tree?: DependencyTree;
}

export function courseName(course: CourseSummary): string {
    return `${course.course_category} ${course.course_code}`;
}

export function formatRequirementKind(kind?: RequirementKind): string {
    switch (kind) {
        case "prerequisite":
            return "Prerequisite";
        case "corequisite":
            return "Corequisite";
        case "pre_or_corequisite":
            return "Pre- or corequisite";
        default:
            return "Course condition";
    }
}
