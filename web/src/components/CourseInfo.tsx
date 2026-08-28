import type {CourseDetailResponse} from "../util";
import {courseName} from "../util";
import PrerequisitePaths from "./PrerequisitePaths";

interface CourseInfoProps {
    detail: CourseDetailResponse;
    onBack: () => void;
}

function CourseInfo({detail, onBack}: CourseInfoProps) {
    const {course, tree} = detail;
    return (
        <article className="course-detail">
            <button className="back-button" type="button" onClick={onBack}>← Back</button>
            <header className="detail-header">
                <div>
                    <p className="section-label">{course.course_category} course</p>
                    <h1><span>{courseName(course)}</span>{course.course_title}</h1>
                </div>
                <div className="credit-block"><strong>{course.course_credit || "—"}</strong><span>credits</span></div>
            </header>

            <section className="prose-block">
                <p className="section-label">Description</p>
                <p>{course.course_description || "No description is included in this snapshot."}</p>
            </section>

            <section className="requirement-source">
                <p className="section-label">Catalog requirements</p>
                <p>{course.course_prerequisites || "No prerequisites are listed for this course."}</p>
            </section>

            <section className="paths-section">
                <div className="section-heading compact">
                    <div>
                        <p className="section-label">Prerequisites</p>
                        <h2>{tree ? "Course requirements" : "No prerequisites listed"}</h2>
                    </div>
                </div>
                {tree ? <PrerequisitePaths initialTree={tree}/> : <div className="empty-state">No course prerequisites listed.</div>}
            </section>
        </article>
    );
}

export default CourseInfo;
