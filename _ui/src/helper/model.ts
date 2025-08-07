export type cell = {
  id: string;
  db_type: string;
  content: string;
  description?: string; // Optional field for description
  collapsed?: boolean;
};

export type notebook = {
  name: string;
  cells: cell[];
};

export type info = {
  version: string;
  databases?: string[];
};
