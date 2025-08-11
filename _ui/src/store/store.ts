import type { idName, info } from "@/helper/model";
import { writable } from "svelte/store";

let navbar = {
  title: "",
  sideBarOpen: true,
};

let noteIds: idName[] = [];

export type QueryOutput = {
  rows_affected?: number;
  rows?: string[][];
  columns: string[];
  duration?: string;
}

export const storeNavbar = writable(navbar);
export const storeInfo = writable<info | null>(null);
export const storeOutput = writable<QueryOutput | null>(null);
export const storeNoteIds = writable<idName[]>(noteIds);
