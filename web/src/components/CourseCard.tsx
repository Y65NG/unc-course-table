import type {CourseSummary} from "../util";
import {courseName} from "../util";

interface CourseCardProps {
    course: CourseSummary;
    onSelect: (course: CourseSummary) => void;
}

function CourseCard({course, onSelect}: CourseCardProps) {
    const creditLabel = course.course_credit
        ? `${course.course_credit} ${course.course_credit === "1" ? "credit" : "credits"}`
        : "Credit varies";

    return (
        <button className="course-card" type="button" onClick={() => onSelect(course)}>
            <span className="course-code">{courseName(course)}</span>
            <span className="course-title"><strong>{course.course_title || "Untitled course"}</strong></span>
            <span className="course-meta">{creditLabel}</span>
            <span className="card-action" aria-hidden="true">→</span>
        </button>
    );
}

export default CourseCard;
