export type cell = {
  id: string;
  db_type: string;
  content: string;
  limit: number;
  template: template;
  mode?: modeTransfer;
  enabled?: boolean;
  result?: boolean;
  description?: string; // Optional field for description
  collapsed?: boolean;
};

export type template = {
  enabled: boolean;
}

export type modeTransfer = {
  enabled: boolean;
  name: "transfer";
  db_type: string;
  table: string;
  batch: number;
  wipe: boolean;
  map_type: map_type;
  skip_error: skip_error;
}

export type skip_error = {
  enabled: boolean;
  message: string;
};

export type map_type = {
  enabled: boolean;
  column?: Record<string, {
    type: "number" | "string";
    nullable: boolean;
  }>;
  destination?: Record<string, {
    type: "number" | "string";
    nullable: boolean;
    template: enabled;
    encoding: encoding;
  }>;
}

export const encodingTypes = ["ISO 8859-1"];

export type encoding = {
  enabled: boolean;
  coding: string;
}

export type enabled = {
  enabled: boolean;
  value: string;
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
