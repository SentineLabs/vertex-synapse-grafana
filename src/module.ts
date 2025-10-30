import { DataSourcePlugin } from '@grafana/data';
import { SynapseCortexDataSource } from './datasource';
import { ConfigEditor } from './components/ConfigEditor';
import { QueryEditor } from './components/QueryEditor';
import { SynapseCortexQuery, SynapseCortexDataSourceOptions } from './types';

export const plugin = new DataSourcePlugin<
  SynapseCortexDataSource,
  SynapseCortexQuery,
  SynapseCortexDataSourceOptions
>(SynapseCortexDataSource)
  .setConfigEditor(ConfigEditor)
  .setQueryEditor(QueryEditor);
