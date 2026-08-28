import type {CourseDetailResponse, SearchResponse} from "./util";

const apiBaseUrl = (import.meta.env.VITE_API_BASE_URL ?? "http://localhost:8080").replace(/\/$/, "");

async function getJSON<T>(path: string, signal?: AbortSignal): Promise<T> {
    const response = await fetch(`${apiBaseUrl}${path}`, {signal});
    if (!response.ok) {
        const body = await response.json().catch(() => null) as {error?: string} | null;
        throw new Error(body?.error ?? `Request failed with status ${response.status}`);
    }
    return response.json() as Promise<T>;
}

export function searchCourses(query: string, page: number, signal?: AbortSignal): Promise<SearchResponse> {
    const params = new URLSearchParams({q: query, limit: "30", page: String(page)});
    return getJSON<SearchResponse>(`/api/courses?${params}`, signal);
}

export function getCourse(subject: string, code: string, depth = 1, signal?: AbortSignal): Promise<CourseDetailResponse> {
    return getJSON<CourseDetailResponse>(
        `/api/courses/${encodeURIComponent(subject)}/${encodeURIComponent(code)}?depth=${depth}`,
        signal,
    );
}
