import { storeDatabases } from "@/store/store";
import axios from "axios";

type RunRequest = {
  name: string;
  query: string;
  args?: Record<string, any>;
}

export const requestRun = (data: RunRequest) => {
  return axios.post("/api/v1/run", data);
};

export const requestDatabases = () => {
  try {
    return axios.get("/api/v1/databases").then(response => {
      const databases = response.data?.data || [];
      storeDatabases.set(databases);
    });
  } catch (error) {
    console.error("Error fetching databases:", error);
  }
};
