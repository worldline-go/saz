import { storeInfo, storeNoteIds } from "@/store/store";
import axios from "axios";
import type { info, notebook, idName, cell } from "./model";
import { addToast } from "@/store/toast";

export const requestRun = (data: cell) => {
  return axios.post("./api/v1/run", data);
};

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

export const requestSave = async (note: notebook) => {
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
