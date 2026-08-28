import type {FormEventHandler} from "react";

interface SearchBoxProps {
    query: string;
    onQueryChange: (value: string) => void;
    onSubmit: FormEventHandler<HTMLFormElement>;
    loading: boolean;
}

function SearchBox({query, onQueryChange, onSubmit, loading}: SearchBoxProps) {
    return (
        <form className="search-box" onSubmit={onSubmit} role="search">
            <label className="sr-only" htmlFor="course-search">Search courses</label>
            <input
                id="course-search"
                value={query}
                onChange={(event) => onQueryChange(event.target.value)}
                placeholder="Search COMP 210 or data structures"
                autoComplete="off"
            />
            <button type="submit" disabled={loading}>{loading ? "Searching" : "Search"}</button>
        </form>
    );
}

export default SearchBox;
