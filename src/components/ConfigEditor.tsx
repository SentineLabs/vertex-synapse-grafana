import React, { ChangeEvent, PureComponent } from 'react';
import { Field, Input, SecretInput, Alert, Switch } from '@grafana/ui';
import { DataSourcePluginOptionsEditorProps, onUpdateDatasourceSecureJsonDataOption } from '@grafana/data';
import { SynapseCortexDataSourceOptions, SynapseCortexSecureJsonData } from '../types';

interface Props extends DataSourcePluginOptionsEditorProps<SynapseCortexDataSourceOptions> {}

interface State {}

export class ConfigEditor extends PureComponent<Props, State> {
  onURLChange = (event: ChangeEvent<HTMLInputElement>) => {
    const { onOptionsChange, options } = this.props;
    console.log(options);
    const jsonData = {
      ...options.jsonData,
      url: event.target.value,
    };
    onOptionsChange({ ...options, url:event.target.value });
  };

  onVersionChange = (event: ChangeEvent<HTMLInputElement>) => {
    const { onOptionsChange, options } = this.props;
    const jsonData = {
      ...options.jsonData,
      version: event.target.value,
    };
    onOptionsChange({ ...options, jsonData });
  };

  onTimeoutChange = (event: ChangeEvent<HTMLInputElement>) => {
    const { onOptionsChange, options } = this.props;
    const jsonData = {
      ...options.jsonData,
      timeout: parseInt(event.target.value, 10),
    };
    onOptionsChange({ ...options, jsonData });
  };

  onTlsSkipVerifyChange = (event: ChangeEvent<HTMLInputElement>) => {
    const { onOptionsChange, options } = this.props;
    const jsonData = {
      ...options.jsonData,
      tlsSkipVerify: event.target.checked,
    };
    onOptionsChange({ ...options, jsonData });
  };


  // Secure field (password) change
  onApiKeyChange = (event: ChangeEvent<HTMLInputElement>) => {
    const { onOptionsChange, options } = this.props;
    console.log(options);
    onOptionsChange({
      ...options,
      secureJsonData: {
        ...options.secureJsonData,
        apiKey: event.target.value,
      },
    });
  };

  onResetApiKey = () => {
    const { onOptionsChange, options } = this.props;
    onOptionsChange({
      ...options,
      secureJsonFields: {
        ...options.secureJsonFields,
        apiKey: false,
      },
      secureJsonData: {
        ...options.secureJsonData,
        apiKey: '',
      },
    });
  };

  render() {
    const { options } = this.props;
    const { url, jsonData, secureJsonFields } = options;
    const secureJsonData = (options.secureJsonData || {}) as SynapseCortexSecureJsonData;

    const apiKeyProps = {
      isConfigured: Boolean(options.secureJsonFields.apiKey),
      value: secureJsonData?.apiKey || '',
      placeholder: 'Enter API key',
      id: 'apiKey',
      onReset: () =>
        this.props.onOptionsChange({
          ...options,
          secureJsonFields: { ...options.secureJsonFields, apiKey: false },
          secureJsonData: { apiKey: '' },
          jsonData: options.jsonData,
        }),
      onChange: onUpdateDatasourceSecureJsonDataOption(this.props, 'apiKey'),
    };

    return (
      <div>
        <Field
          label="URL"
          description="The URL of your Vertex Synapse Cortex instance"
          required
        >
          <Input
            onChange={this.onURLChange}
            value={url || ''}
            placeholder="https://your-cortex-instance.com"
            width={40}
          />
        </Field>

        <Field
          label="API Version"
          description="API version to use (default: v1)"
        >
          <Input
            onChange={this.onVersionChange}
            value={jsonData.version || 'v1'}
            placeholder="v1"
            width={20}
          />
        </Field>

        <Field
          label="Timeout (ms)"
          description="Request timeout in milliseconds"
        >
          <Input
            onChange={this.onTimeoutChange}
            value={jsonData.timeout || 30000}
            placeholder="30000"
            type="number"
            width={20}
          />
        </Field>

        <Field
          label="Skip TLS Verification"
          description="Skip TLS certificate verification (useful for self-signed certificates)"
        >
          <Switch
            value={options.jsonData.tlsSkipVerify || false}
            onChange={this.onTlsSkipVerifyChange}
          />
        </Field>

        <Field
          label="API Key"
          description="Your Synapse Cortex API key"
        >
          <SecretInput
            {...apiKeyProps} label="API key" width={40} 
          />
        </Field>

        <Alert title="Vertex Synapse Cortex Configuration" severity="info">
          <p>
            This datasource connects to Vertex Synapse Cortex, a hypergraph analysis platform. 
            Configure the URL of your Cortex instance and choose your authentication method.
          </p>
          <p>
            <strong>Supported Query Types:</strong>
          </p>
          <ul>
            <li><strong>Nodes:</strong> Query nodes by form and properties</li>
            <li><strong>Edges:</strong> Query edges and relationships</li>
            <li><strong>Storm:</strong> Execute Storm query language</li>
            <li><strong>Search:</strong> Full-text search across the hypergraph</li>
          </ul>
        </Alert>
      </div>
    );
  }
}
