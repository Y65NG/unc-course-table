import {FormEvent, useEffect, useRef, useState} from "react";
import "./App.scss";
import {getCourse, searchCourses} from "./api";
import CourseBoard from "./components/CourseBoard";
import CourseInfo from "./components/CourseInfo";
import Pagination from "./components/Pagination";
import SearchBox from "./components/SearchBox";
import type {CourseDetailResponse, CourseSummary} from "./util";

function App() {
    const [query, setQuery] = useState("COMP");
    const [submittedQuery, setSubmittedQuery] = useState("COMP");
    const [courses, setCourses] = useState<CourseSummary[]>([]);
    const [page, setPage] = useState(1);
    const [total, setTotal] = useState(0);
    const [totalPages, setTotalPages] = useState(0);
    const [selectedCourse, setSelectedCourse] = useState<CourseDetailResponse | null>(null);
    const [loading, setLoading] = useState(false);
    const [error, setError] = useState("");
    const requestId = useRef(0);

    useEffect(() => {
        const controller = new AbortController();
        const currentRequest = ++requestId.current;
        setLoading(true);
        setError("");
        searchCourses(submittedQuery, page, controller.signal)
            .then((response) => {
                if (currentRequest === requestId.current) {
                    setCourses(response.results);
                    setTotal(response.total);
                    setTotalPages(response.total_pages);
                }
            })
            .catch((reason: unknown) => {
                if (reason instanceof DOMException && reason.name === "AbortError") return;
                if (currentRequest === requestId.current) {
                    setCourses([]);
                    setTotal(0);
                    setTotalPages(0);
                    setError(reason instanceof Error ? reason.message : "Unable to search the catalog.");
                }
            })
            .finally(() => {
                if (currentRequest === requestId.current) setLoading(false);
            });
        return () => controller.abort();
    }, [page, submittedQuery]);

    function submitSearch(event: FormEvent) {
        event.preventDefault();
        setSelectedCourse(null);
        setPage(1);
        setSubmittedQuery(query.trim());
    }

    function changePage(nextPage: number) {
        setPage(nextPage);
        document.querySelector(".results-section")?.scrollIntoView({block: "start"});
    }

    async function selectCourse(course: CourseSummary) {
        setLoading(true);
        setError("");
        try {
            const detail = await getCourse(course.course_category, course.course_code);
            setSelectedCourse(detail);
            window.scrollTo({top: 0, behavior: "smooth"});
        } catch (reason) {
            setError(reason instanceof Error ? reason.message : "Unable to load the course.");
        } finally {
            setLoading(false);
        }
    }

    return (
        <div className="app-shell">
            <header className="site-header">
                <a className="wordmark" href="/" aria-label="UNC Course Explorer home">
                    UNC Course Prerequisite Tool
                </a>
                <a className="header-link" href="https://catalog.unc.edu/courses/" target="_blank" rel="noreferrer">
                    UNC Catalog
                </a>
            </header>

            <main>
                {!selectedCourse && (
                    <section className="search-intro">
                        <div className="intro-copy">
                            <h1>Search UNC courses</h1>
                        </div>
                        <div className="search-area">
                            <SearchBox query={query} onQueryChange={setQuery} onSubmit={submitSearch} loading={loading}/>
                        </div>
                    </section>
                )}

                {error && <div className="notice notice-error" role="alert">{error}</div>}

                {selectedCourse ? (
                    <CourseInfo detail={selectedCourse} onBack={() => setSelectedCourse(null)}/>
                ) : (
                    <section className="results-section" aria-live="polite">
                        <div className="section-heading">
                            <h2>{loading ? "Searching…" : `${total} results for “${submittedQuery || "all courses"}”`}</h2>
                        </div>
                        {!loading && courses.length === 0 && !error ? (
                            <div className="empty-state">No courses found. Try a subject code, number, or title.</div>
                        ) : (
                            <>
                                <CourseBoard courses={courses} onSelect={selectCourse}/>
                                <Pagination currentPage={page} totalPages={totalPages} disabled={loading} onPageChange={changePage}/>
                            </>
                        )}
                    </section>
                )}
            </main>

        </div>
    );
}

export default App;
