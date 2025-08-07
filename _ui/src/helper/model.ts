export type cell = {
  id: string;
  db_type: string;
  content: string;
  description?: string; // Optional field for description
  collapsed?: boolean;
};

export type notebook = {
  id: string;
  name: string;
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
