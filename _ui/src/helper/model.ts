export type cell = {
  id: string;
  db_type: string;
  content: string;
  limit: number;
  mode?: modeTransfer;
  enabled?: boolean;
  result?: boolean;
  description?: string; // Optional field for description
  collapsed?: boolean;
};

export type modeTransfer = {
  enabled: boolean;
  name: "transfer";
  db_type: string;
  table: string;
  wipe: boolean;
}

export type notebook = {
  id: string;
  name: string;
  path: string;
  content: content;
};

export type content = {
  cells: cell[];
}

export type info = {
  version: string;
  databases?: string[];
};

export type idName = {
  id: string;
  name: string;
};
