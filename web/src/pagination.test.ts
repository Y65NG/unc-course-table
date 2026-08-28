import {describe, expect, it} from "vitest";
import {paginationItems} from "./pagination";

describe("paginationItems", () => {
    it("shows every page for short result sets", () => {
        expect(paginationItems(6, 1)).toEqual([1, 2, 3, 4, 5, 6]);
    });

    it("keeps the first, current, and last pages for long result sets", () => {
        expect(paginationItems(20, 10)).toEqual([1, 2, 3, "ellipsis", 9, 10, 11, "ellipsis", 18, 19, 20]);
    });

    it("uses one ellipsis when the current page is near an edge", () => {
        expect(paginationItems(20, 2)).toEqual([1, 2, 3, "ellipsis", 18, 19, 20]);
    });
});
