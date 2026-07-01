import * as React from 'react';
import { LogViewer } from './LogViewer';

export interface DeploymentLogsTabProps {
  appId: string;
}

export function DeploymentLogsTab({ appId }: DeploymentLogsTabProps) {
  return <LogViewer appId={appId} footerLabel="Deployment #" />;
}
