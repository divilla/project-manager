export interface Epic {
  id: number;
  version: number;
  project_id: number;
  name: string;
  done_req: number;
  total_req: number;
  completed: number;
  change_count: number;
  created: string;
  modified: string;
}

export interface EpicCreateInput {
  project_id: number;
  name: string;
}

export interface EpicUpdateInput {
  id: number;
  name: string;
}
