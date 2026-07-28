import { useEffect } from 'react';
import { authApi } from '../../api/endpoints';
import type { LogMessage } from '../../types';

interface LogStreamerProps {
  deploymentId: string;
  onLog: (line: string) => void;
  onStatusChange: (status: string) => void;
}

function websocketUrl(deploymentId: string, ticket: string) {
  const protocol = window.location.protocol === 'https:' ? 'wss:' : 'ws:';
  return `${protocol}//${window.location.host}/api/v1/deployments/${deploymentId}/ws?ticket=${encodeURIComponent(ticket)}`;
}

export function LogStreamer({ deploymentId, onLog, onStatusChange }: LogStreamerProps) {
  useEffect(() => {
    let socket: WebSocket | undefined;
    let cancelled = false;

    onStatusChange('connecting');

    authApi.wsTicket()
      .then(({ ticket }) => {
        if (cancelled) {
          return;
        }

        socket = new WebSocket(websocketUrl(deploymentId, ticket));

        socket.onopen = () => onStatusChange('connected');
        socket.onmessage = (event) => {
          const message = JSON.parse(event.data) as LogMessage;
          onLog(message.line);
        };
        socket.onerror = () => onStatusChange('error');
        socket.onclose = () => onStatusChange('closed');
      })
      .catch((error: Error) => {
        onStatusChange('error');
        onLog(`Failed to connect to log stream: ${error.message}`);
      });

    return () => {
      cancelled = true;
      socket?.close();
    };
  }, [deploymentId, onLog, onStatusChange]);

  return null;
}
