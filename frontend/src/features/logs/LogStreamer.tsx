import { useEffect, useRef } from 'react';
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
  // keep stable refs to callbacks so useEffect doesn't re-fire on each render
  const onLogRef = useRef(onLog);
  const onStatusRef = useRef(onStatusChange);
  useEffect(() => { onLogRef.current = onLog; }, [onLog]);
  useEffect(() => { onStatusRef.current = onStatusChange; }, [onStatusChange]);

  useEffect(() => {
    let socket: WebSocket | undefined;
    let cancelled = false;

    onStatusRef.current('connecting');

    // small delay to let React StrictMode's unmount/remount cycle settle
    // before we consume the one-time ws ticket from Redis
    const timer = setTimeout(() => {
      if (cancelled) return;

      authApi.wsTicket()
        .then(({ ticket }) => {
          if (cancelled) return;

          const url = websocketUrl(deploymentId, ticket);
          socket = new WebSocket(url);

          socket.onopen = () => onStatusRef.current('connected');

          socket.onmessage = (event) => {
            try {
              const message = JSON.parse(event.data) as LogMessage;
              onLogRef.current(message.line);
            } catch {
              // ignore malformed frames
            }
          };

          socket.onerror = () => onStatusRef.current('error');
          socket.onclose = () => onStatusRef.current('closed');
        })
        .catch((error: Error) => {
          if (cancelled) return;
          onStatusRef.current('error');
          onLogRef.current(`Failed to connect to log stream: ${error.message}`);
        });
    }, 100); // 100 ms debounce beats StrictMode double-invoke

    return () => {
      cancelled = true;
      clearTimeout(timer);
      socket?.close();
    };
  }, [deploymentId]); // only re-run when the deployment actually changes

  return null;
}
