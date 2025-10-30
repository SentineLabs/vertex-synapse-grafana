import { DataQuery, DataSourceJsonData } from '@grafana/data';

export interface SynapseCortexQuery extends DataQuery {
  stormQuery: string;
  useCall?: boolean;
  opts?: Record<string, any>;
}

export const DEFAULT_QUERY: Partial<SynapseCortexQuery> = {
  stormQuery: '',
  useCall: false,
  opts: {},
};

export interface SynapseCortexDataSourceOptions extends DataSourceJsonData {
  url?: string;
  version?: string;
  timeout?: number;
  tlsSkipVerify?: boolean;
}

export interface SynapseCortexSecureJsonData {
  apiKey?: string;
}

// API Response Types
export interface CortexNode {
  ndef: [string, any]; // [form, props]
  tags?: Record<string, any>;
  tagprops?: Record<string, any>;
  path?: {
    iden: string;
    tags?: Record<string, any>;
  };
}

export interface CortexEdge {
  ndef: [string, any]; // [form, props]
  n1ndef: [string, any];
  n2ndef: [string, any];
}

export interface CortexApiResponse {
  result?: CortexNode[] | CortexEdge[] | any[];
  error?: {
    code: string;
    mesg: string;
  };
  took?: number;
}

export interface StormQueryResponse {
  result?: any[];
  warnings?: string[];
  error?: {
    code: string;
    mesg: string;
  };
  took?: number;
}
