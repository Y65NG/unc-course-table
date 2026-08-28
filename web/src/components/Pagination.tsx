import {paginationItems} from "../pagination";

interface PaginationProps {
    currentPage: number;
    totalPages: number;
    disabled: boolean;
    onPageChange: (page: number) => void;
}

export default function Pagination({currentPage, totalPages, disabled, onPageChange}: PaginationProps) {
    if (totalPages <= 1) return null;

    return (
        <nav className="pagination" aria-label="Search result pages">
            {paginationItems(totalPages, currentPage).map((item, index) => item === "ellipsis" ? (
                <span className="pagination-ellipsis" aria-hidden="true" key={`ellipsis-${index}`}>…</span>
            ) : (
                <button
                    type="button"
                    className={item === currentPage ? "is-current" : ""}
                    aria-current={item === currentPage ? "page" : undefined}
                    aria-label={`Page ${item}`}
                    disabled={disabled || item === currentPage}
                    onClick={() => onPageChange(item)}
                    key={item}
                >
                    {item}
                </button>
            ))}
        </nav>
    );
}
