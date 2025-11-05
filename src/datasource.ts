import {
  DataSourceInstanceSettings,
} from '@grafana/data';

import { DataSourceWithBackend } from '@grafana/runtime';

import { SynapseCortexQuery, SynapseCortexDataSourceOptions } from './types';

export class SynapseCortexDataSource extends DataSourceWithBackend<SynapseCortexQuery, SynapseCortexDataSourceOptions> {
  constructor(instanceSettings: DataSourceInstanceSettings<SynapseCortexDataSourceOptions>) {
    super(instanceSettings);
  }

}
