import axios from "axios";

type RunRequest = {
  name: string;
  query: string;
  args?: Record<string, any>;
}

export const requestRun = (data: RunRequest) => {
  return axios.post("/api/v1/run", data);
};
