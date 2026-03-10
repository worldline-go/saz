import { storeInfo, storeNoteIds } from "@/store/store";
import axios from "axios";
import type { info, notebook, idName, cellPlus, process } from "./model";
import { addToast } from "@/store/toast";

export const requestRun = (data: cellPlus) => {
  return axios.post("./api/v1/run", data);
};

export const requestRunBackground = (data: cellPlus) => {
  return axios.post("./api/v1/run/background", data);
};

export const requestRunTemplate = (content: string, data: any) => {
  return axios.post("./api/v1/render", { content, data });
}

export const requestRunNotebook = (path: string) => {
  return axios.post(`./api/v1/run/${path}`);
};

export const requestInfo = async () => {
  try {
    const response = await axios.get("./api/v1/info");
    const info = response.data?.data as info;
    storeInfo.set(info);
  } catch (error) {
    addToast("Error fetching info", "alert");
    console.error("Error fetching info:", error);
  }
};

type NotesResponse = {
  data: idName[];
};

type NoteResponse = {
  data?: notebook;
};

export const requestNotes = async () => {
  try {
    const response = await axios.get("./api/v1/notes");
    const notes = response.data as NotesResponse;
    storeNoteIds.set(notes.data);
  } catch (error) {
    addToast("Error fetching notes", "alert");
    console.error("Error fetching notes:", error);
  }
};

export const requestNote = async (id: string) => {
  try {
    const response = await axios.get(`./api/v1/notes/${id}`);
    const note = response.data as NoteResponse;
    return note.data;
  } catch (error) {
    addToast("Error fetching note", "alert");
    console.error("Error fetching notes:", error);
  }
};

export const requestNoteDelete = (id: string) => {
  return axios.delete(`./api/v1/notes/${id}`);
};

type ProcessResponse = {
  data: process[];
};

export type ProcessQueryParams = {
  sort?: string;
  limit?: number;
  offset?: number;
  status?: string;
};

/**
 * Build a rakunlabs/query-compatible query string from params.
 * e.g. ?_sort=-created_at&_limit=25&_offset=0&status=running
 */
function buildProcessQuery(params?: ProcessQueryParams): string {
  if (!params) return "";
  const parts: string[] = [];
  if (params.sort) parts.push(`_sort=${encodeURIComponent(params.sort)}`);
  if (params.limit != null) parts.push(`_limit=${params.limit}`);
  if (params.offset != null) parts.push(`_offset=${params.offset}`);
  if (params.status) parts.push(`status=${encodeURIComponent(params.status)}`);
  return parts.length ? `?${parts.join("&")}` : "";
}

export const requestProcesses = async (params?: ProcessQueryParams): Promise<process[]> => {
  try {
    const qs = buildProcessQuery(params);
    const response = await axios.get(`./api/v1/process${qs}`);
    const processes = response.data as ProcessResponse;
    return processes.data || [];
  } catch (error) {
    addToast("Error fetching processes", "alert");
    console.error("Error fetching processes:", error);
    return [];
  }
};

export const requestProcessTerminate = async (pid: string) => {
  return axios.post(`./api/v1/process/${pid}`, { action: "terminate" });
};

export const requestProcessDelete = async () => {
  return axios.delete("./api/v1/process");
};

export const requestNoteSave = async (note: notebook) => {
  try {
    await axios.put(`./api/v1/notes/${note.id}`, note)
    storeNoteIds.update((ids) => {
      const index = ids.findIndex((n) => n.id === note.id);
      if (index !== -1) {
        ids[index] = { id: note.id, name: note.name };
      } else {
        ids.push({ id: note.id, name: note.name });
      }

      return ids;
    });
    addToast(`Note "${note.name}" saved successfully!`, "info");
  } catch (error) {
    addToast("Error saving note", "alert");
    console.error("Error saving note:", error);
  }
}
