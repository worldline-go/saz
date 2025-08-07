import type { info } from "@/helper/model";
import { writable } from "svelte/store";

let navbar = {
  title: "",
  sideBarOpen: true,
};

export const storeNavbar = writable(navbar);
export const storeInfo = writable<info | null>(null);
export const storeOutput = writable<Record<string, any>[]>([]);
