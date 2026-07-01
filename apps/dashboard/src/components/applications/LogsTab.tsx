import * as React from 'react';
import { LogViewer } from './LogViewer';

export interface LogsTabProps {
  appId: string;
}

export function LogsTab({ appId }: LogsTabProps) {
  return <LogViewer appId={appId} footerLabel={appId} />;
}
