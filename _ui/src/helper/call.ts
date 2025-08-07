import { storeInfo } from "@/store/store";
import axios from "axios";
import type { info } from "./model";

type RunRequest = {
  name: string;
  query: string;
  args?: Record<string, any>;
}

export const requestRun = (data: RunRequest) => {
  return axios.post("/api/v1/run", data);
};

export const requestInfo = () => {
  try {
    return axios.get("/api/v1/info").then(response => {
      const info = response.data?.data as info;
      storeInfo.set(info);
    });
  } catch (error) {
    console.error("Error fetching info:", error);
  }
};
