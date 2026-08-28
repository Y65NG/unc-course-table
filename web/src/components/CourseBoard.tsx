import type {CourseSummary} from "../util";
import CourseCard from "./CourseCard";

interface CourseBoardProps {
    courses: CourseSummary[];
    onSelect: (course: CourseSummary) => void;
}

function CourseBoard({courses, onSelect}: CourseBoardProps) {
    return (
        <div className="course-list">
            {courses.map((course) => (
                <CourseCard key={`${course.course_category}-${course.course_code}`} course={course} onSelect={onSelect}/>
            ))}
        </div>
    );
}

export default CourseBoard;
