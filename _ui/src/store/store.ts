import type { idName, info } from "@/helper/model";
import { writable } from "svelte/store";

let navbar = {
  title: "",
  sideBarOpen: true,
};

let noteIds: idName[] = [];

export const storeNavbar = writable(navbar);
export const storeInfo = writable<info | null>(null);
export const storeOutput = writable<Record<string, any>[]>([]);
export const storeNoteIds = writable<idName[]>(noteIds);
