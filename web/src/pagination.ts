export type PaginationItem = number | "ellipsis";

export function paginationItems(totalPages: number, currentPage: number): PaginationItem[] {
    if (totalPages <= 0) return [];
    if (totalPages <= 10) return Array.from({length: totalPages}, (_, index) => index + 1);

    const pages = new Set<number>();
    for (const page of [1, 2, 3, currentPage - 1, currentPage, currentPage + 1, totalPages - 2, totalPages - 1, totalPages]) {
        if (page >= 1 && page <= totalPages) pages.add(page);
    }

    const sorted = [...pages].sort((left, right) => left - right);
    const items: PaginationItem[] = [];
    for (const page of sorted) {
        const previous = items[items.length - 1];
        if (typeof previous === "number") {
            if (page - previous === 2) items.push(previous + 1);
            if (page - previous > 2) items.push("ellipsis");
        }
        items.push(page);
    }
    return items;
}
