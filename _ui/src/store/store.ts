import { writable } from "svelte/store";

let navbar = {
  title: "",
  sideBarOpen: true,
};

export const storeNavbar = writable(navbar);
export const storeDatabases = writable<string[]>([]);
export const storeOutput = writable<Record<string, any>[]>([]);
