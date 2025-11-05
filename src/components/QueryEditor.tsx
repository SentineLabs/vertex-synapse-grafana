import React, { ChangeEvent } from 'react';
import { InlineField, Input, InlineFieldRow, InlineSwitch, TextArea, Select } from '@grafana/ui';
import { QueryEditorProps, SelectableValue } from '@grafana/data';
import { SynapseCortexDataSource } from '../datasource';
import { SynapseCortexDataSourceOptions, SynapseCortexQuery } from '../types';

type Props = QueryEditorProps<SynapseCortexDataSource, SynapseCortexQuery, SynapseCortexDataSourceOptions>;

export function QueryEditor({ query, onChange, onRunQuery }: Props) {
  const onStormQueryChange = (event: ChangeEvent<HTMLTextAreaElement>) => {
    onChange({ ...query, stormQuery: event.target.value });
  };

  const onUseCallChange = (event: ChangeEvent<HTMLInputElement>) => {
    onChange({ ...query, useCall: event.currentTarget.checked });
  };

  const onLimitChange = (event: ChangeEvent<HTMLInputElement>) => {
    const limit = parseInt(event.target.value, 10);
    onChange({ 
      ...query, 
      opts: { 
        ...query.opts, 
        limit: isNaN(limit) ? undefined : limit 
      } 
    });
  };

  const onReadonlyChange = (event: ChangeEvent<HTMLInputElement>) => {
    onChange({ 
      ...query, 
      opts: { 
        ...query.opts, 
        readonly: event.currentTarget.checked 
      } 
    });
  };

  const onReprChange = (event: ChangeEvent<HTMLInputElement>) => {
    onChange({ 
      ...query, 
      opts: { 
        ...query.opts, 
        repr: event.currentTarget.checked 
      } 
    });
  };

  const onFlattenChange = (event: ChangeEvent<HTMLInputElement>) => {
    onChange({ 
      ...query, 
      opts: { 
        ...query.opts, 
        flatten: event.currentTarget.checked 
      } 
    });
  };

  const onEditFormatChange = (value: SelectableValue<string>) => {
    onChange({ 
      ...query, 
      opts: { 
        ...query.opts, 
        editformat: value?.value || undefined 
      } 
    });
  };

  const onModeChange = (value: SelectableValue<string>) => {
    onChange({ 
      ...query, 
      opts: { 
        ...query.opts, 
        mode: value?.value || undefined 
      } 
    });
  };

  const onPathChange = (event: ChangeEvent<HTMLInputElement>) => {
    onChange({ 
      ...query, 
      opts: { 
        ...query.opts, 
        path: event.currentTarget.checked 
      } 
    });
  };

  const onLinksChange = (event: ChangeEvent<HTMLInputElement>) => {
    onChange({ 
      ...query, 
      opts: { 
        ...query.opts, 
        links: event.currentTarget.checked 
      } 
    });
  };

  const onVarsChange = (event: ChangeEvent<HTMLInputElement>) => {
    try {
      const vars = event.target.value ? JSON.parse(event.target.value) : undefined;
      onChange({ 
        ...query, 
        opts: { 
          ...query.opts, 
          vars
        } 
      });
    } catch (e) {
      // Invalid JSON, ignore for now
    }
  };

  const onViewChange = (event: ChangeEvent<HTMLInputElement>) => {
    onChange({ 
      ...query, 
      opts: { 
        ...query.opts, 
        view: event.target.value || undefined 
      } 
    });
  };

  const onKeyDown = (event: React.KeyboardEvent) => {
    if (event.key === 'Enter' && (event.ctrlKey || event.metaKey)) {
      onRunQuery();
    }
  };

  return (
    <div>
      <InlineFieldRow>
        <InlineField label="Storm Query" labelWidth={16} grow>
          <TextArea
            value={query.stormQuery || ''}
            onChange={onStormQueryChange}
            onKeyDown={onKeyDown}
            placeholder="Enter Storm query, e.g., inet:fqdn +#malware | limit 10"
            rows={4}
            style={{ width: '100%' }}
          />
        </InlineField>
      </InlineFieldRow>

      <InlineFieldRow>
        <InlineField label="Use Call API" labelWidth={16} tooltip="Use the Storm call API endpoint for single-result queries">
          <InlineSwitch
            value={query.useCall || false}
            onChange={onUseCallChange}
          />
        </InlineField>
        {query.useCall && (
          <InlineField label="Flatten Nested" tooltip="Flatten nested objects into dot-notation columns (e.g., data.edits:meta.total)">
            <InlineSwitch
              value={query.opts?.flatten || false}
              onChange={onFlattenChange}
            />
          </InlineField>
        )}
      </InlineFieldRow>

      <InlineFieldRow>
        <InlineField label="Limit" labelWidth={16} tooltip="Maximum number of nodes to return (for streaming queries)">
          <Input
            value={query.opts?.limit || ''}
            onChange={onLimitChange}
            onKeyDown={onKeyDown}
            placeholder="100"
            type="number"
            width={16}
          />
        </InlineField>
        <InlineField label="Read Only" tooltip="Execute query in read-only mode">
          <InlineSwitch
            value={query.opts?.readonly !== false}
            onChange={onReadonlyChange}
          />
        </InlineField>
      </InlineFieldRow>

      <InlineFieldRow>
        <InlineField label="Mode" labelWidth={16} tooltip="Query parsing mode: lookup, autoadd, or search">
          <Select
            value={query.opts?.mode}
            onChange={onModeChange}
            options={[
              { label: 'Default', value: '' },
              { label: 'Lookup', value: 'lookup' },
              { label: 'Auto Add', value: 'autoadd' },
              { label: 'Search', value: 'search' },
            ]}
            width={24}
            isClearable
          />
        </InlineField>
        <InlineField label="Edit Format" tooltip="Format for node edits: nodeedits, none, or count">
          <Select
            value={query.opts?.editformat}
            onChange={onEditFormatChange}
            options={[
              { label: 'Node Edits', value: 'nodeedits' },
              { label: 'None', value: 'none' },
              { label: 'Count', value: 'count' },
            ]}
            width={24}
            isClearable
          />
        </InlineField>
      </InlineFieldRow>

      <InlineFieldRow>
        <InlineField label="Path Info" labelWidth={16} tooltip="Include node path information in results">
          <InlineSwitch
            value={query.opts?.path || false}
            onChange={onPathChange}
          />
        </InlineField>
        <InlineField label="Links" tooltip="Include pivot/edge walk links in results">
          <InlineSwitch
            value={query.opts?.links || false}
            onChange={onLinksChange}
          />
        </InlineField>
        <InlineField label="Repr" tooltip="Include human-friendly value representations">
          <InlineSwitch
            value={query.opts?.repr || false}
            onChange={onReprChange}
          />
        </InlineField>
      </InlineFieldRow>

      <InlineFieldRow>
        <InlineField label="View" labelWidth={16} tooltip="View iden to run the query in (leave empty for default view)">
          <Input
            value={query.opts?.view || ''}
            onChange={onViewChange}
            onKeyDown={onKeyDown}
            placeholder="31ded629eea3c7221be0a61695862952"
            width={40}
          />
        </InlineField>
      </InlineFieldRow>

      <InlineFieldRow>
        <InlineField label="Variables" labelWidth={16} tooltip="JSON object with Storm variables to pass to the query" grow>
          <Input
            value={query.opts?.vars ? JSON.stringify(query.opts.vars) : ''}
            onChange={onVarsChange}
            onKeyDown={onKeyDown}
            placeholder='{"key": "value"}'
            width={48}
          />
        </InlineField>
      </InlineFieldRow>

      <div style={{ 
        marginTop: '10px', 
        padding: '10px', 
        backgroundColor: 'var(--in-content-button-background, rgba(128, 128, 128, 0.1))', 
        borderRadius: '4px',
        border: '1px solid var(--in-content-box-border-color, rgba(128, 128, 128, 0.2))'
      }}>
        <div style={{ fontSize: '12px', color: 'var(--text-primary)' }}>
          <strong>Time variables auto-injected from Grafana picker:</strong>
          <div style={{ marginTop: '4px', lineHeight: '1.5' }}>
            • <code style={{ backgroundColor: 'var(--in-content-button-background-hover, rgba(128, 128, 128, 0.15))', padding: '1px 4px', borderRadius: '2px' }}>$timeRange</code> — Use with @= operator: <code style={{ backgroundColor: 'var(--in-content-button-background-hover, rgba(128, 128, 128, 0.15))', padding: '1px 4px', borderRadius: '2px' }}>.created@=$timeRange</code><br/>
            • <code style={{ backgroundColor: 'var(--in-content-button-background-hover, rgba(128, 128, 128, 0.15))', padding: '1px 4px', borderRadius: '2px' }}>$timeFrom</code>, <code style={{ backgroundColor: 'var(--in-content-button-background-hover, rgba(128, 128, 128, 0.15))', padding: '1px 4px', borderRadius: '2px' }}>$timeTo</code> — ISO 8601 strings<br/>
            • <code style={{ backgroundColor: 'var(--in-content-button-background-hover, rgba(128, 128, 128, 0.15))', padding: '1px 4px', borderRadius: '2px' }}>$dateFrom</code>, <code style={{ backgroundColor: 'var(--in-content-button-background-hover, rgba(128, 128, 128, 0.15))', padding: '1px 4px', borderRadius: '2px' }}>$dateTo</code> — YYYY-MM-DD format
          </div>
          <div style={{ marginTop: '8px' }}>
            <strong>Example:</strong> <code style={{ backgroundColor: 'var(--in-content-button-background-hover, rgba(128, 128, 128, 0.15))', padding: '2px 6px', borderRadius: '2px' }}>inet:flow +:time@=$timeRange | limit 100</code>
          </div>
        </div>
      </div>
    </div>
  );
}
