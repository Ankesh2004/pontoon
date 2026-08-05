import { useEffect, useRef, useState } from 'react';
import { Terminal } from '@xterm/xterm';
import { FitAddon } from '@xterm/addon-fit';
import { WebglAddon } from '@xterm/addon-webgl';
import { ChevronDown } from 'lucide-react';
import { authApi, deploymentsApi } from '../../api/endpoints';
import type { LogMessage } from '../../types';
import '@xterm/xterm/css/xterm.css';

interface TerminalViewProps {
  deploymentId: string;
  status: string;
  buildLogs?: string;
  onConnectionChange?: (status: string) => void;
}

export function TerminalView({ deploymentId, status, buildLogs, onConnectionChange }: TerminalViewProps) {
  const containerRef = useRef<HTMLDivElement>(null);
  const terminalRef = useRef<Terminal | null>(null);
  const fitAddonRef = useRef<FitAddon | null>(null);
  const [isDetached, setIsDetached] = useState(false);
  const scrollTimeoutRef = useRef<number | null>(null);
  const isDetachedRef = useRef(false);

  useEffect(() => {
    isDetachedRef.current = isDetached;
  }, [isDetached]);

  useEffect(() => {
    if (!containerRef.current) return;

    // Initialize Terminal
    const term = new Terminal({
      scrollback: 5000,
      fontFamily: 'Menlo, Monaco, "Courier New", monospace',
      fontSize: 13,
      lineHeight: 1.4,
      theme: {
        background: 'transparent',
        foreground: '#e2e8f0', // slate-200
        cursor: '#3b82f6',     // blue-500
        black: '#000000',
        red: '#ef4444',
        green: '#22c55e',
        yellow: '#eab308',
        blue: '#3b82f6',
        magenta: '#d946ef',
        cyan: '#06b6d4',
        white: '#ffffff',
      },
    });

    const fitAddon = new FitAddon();
    term.loadAddon(fitAddon);

    term.open(containerRef.current);
    
    // Try to load WebGL Addon for performance
    try {
      const webglAddon = new WebglAddon();
      webglAddon.onContextLoss(() => webglAddon.dispose());
      term.loadAddon(webglAddon);
    } catch (e) {
      console.warn("WebGL addon failed to load, falling back to canvas rendering.", e);
    }

    fitAddon.fit();
    terminalRef.current = term;
    fitAddonRef.current = fitAddon;

    // Handle Resize Observer
    const resizeObserver = new ResizeObserver(() => {
      try {
        fitAddon.fit();
      } catch (e) {
        // ignore fit errors when hidden
      }
    });
    resizeObserver.observe(containerRef.current);

    // Track user scrolling
    term.onScroll((newY) => {
      const isAtBottom = newY >= term.buffer.active.baseY - 1; 
      setIsDetached(!isAtBottom);
    });

    return () => {
      resizeObserver.disconnect();
      term.dispose();
      terminalRef.current = null;
      fitAddonRef.current = null;
    };
  }, []);

  // Handle data fetching (WebSocket vs Static)
  useEffect(() => {
    const term = terminalRef.current;
    if (!term) return;

    let socket: WebSocket | undefined;
    let cancelled = false;
    let pingInterval: number | undefined;

    const isActive = ['pending', 'cloning', 'building', 'running'].includes(status.toLowerCase());

    const scrollIfNeeded = () => {
      if (!cancelled && !isDetachedRef.current && terminalRef.current) {
        if (scrollTimeoutRef.current) clearTimeout(scrollTimeoutRef.current);
        scrollTimeoutRef.current = setTimeout(() => {
          terminalRef.current?.scrollToBottom();
        }, 50);
      }
    };

    if (isActive) {
      onConnectionChange?.('connecting');
      const timer = setTimeout(() => {
        if (cancelled) return;

        authApi.wsTicket()
          .then(({ ticket }) => {
            if (cancelled) return;
            const apiBase = import.meta.env.VITE_API_URL || '';
            const protocol = window.location.protocol === 'https:' ? 'wss:' : 'ws:';
            const wsBase = apiBase 
              ? apiBase.replace('http', 'ws') 
              : `${protocol}//api.34.139.115.65.nip.io`; // Hardcoded fallback for Vercel Proxy
              
            const url = `${wsBase}/api/v1/deployments/${deploymentId}/ws?ticket=${encodeURIComponent(ticket)}`;
            
            socket = new WebSocket(url);

            socket.onopen = () => {
              onConnectionChange?.('connected');
              pingInterval = setInterval(() => {
                if (socket?.readyState === WebSocket.OPEN) {
                  socket.send('ping');
                }
              }, 30000);
            };

            socket.onmessage = (event) => {
              try {
                const message = JSON.parse(event.data) as LogMessage;
                term.writeln(message.line);
                scrollIfNeeded();
              } catch {
                term.writeln(event.data);
                scrollIfNeeded();
              }
            };

            socket.onerror = () => onConnectionChange?.('error');
            socket.onclose = () => {
              onConnectionChange?.('closed');
              clearInterval(pingInterval);
            };
          })
          .catch((error) => {
            if (cancelled) return;
            onConnectionChange?.('error');
            term.writeln(`\x1b[31mFailed to connect to log stream: ${error.message}\x1b[0m`);
          });
      }, 100);

      return () => {
        cancelled = true;
        clearTimeout(timer);
        clearInterval(pingInterval);
        socket?.close();
      };
    } else {
      term.clear();
      onConnectionChange?.('closed');
      
      if (buildLogs) {
        term.writeln('\x1b[1;30m=========================================\x1b[0m');
        term.writeln('\x1b[1;36m              BUILD LOGS                 \x1b[0m');
        term.writeln('\x1b[1;30m=========================================\x1b[0m\r\n');
        term.write(buildLogs.replace(/\r\n/g, '\n').replace(/\n/g, '\r\n'));
        term.writeln('\r\n');
      }

      term.writeln('\x1b[1;30m=========================================\x1b[0m');
      term.writeln('\x1b[1;36m             RUNTIME LOGS                \x1b[0m');
      term.writeln('\x1b[1;30m=========================================\x1b[0m\r\n');
      
      deploymentsApi.getLogs(deploymentId)
        .then((logs) => {
          if (cancelled) return;
          if (!logs) {
            term.writeln('\x1b[90mNo runtime logs available.\x1b[0m');
            return;
          }
          // Fix newlines for xterm
          term.write(logs.replace(/\r\n/g, '\n').replace(/\n/g, '\r\n'));
          setTimeout(() => term.scrollToBottom(), 100);
        })
        .catch((err) => {
          if (cancelled) return;
          term.writeln(`\x1b[31mFailed to load historical logs: ${err.message}\x1b[0m`);
        });

      return () => {
        cancelled = true;
      };
    }
  }, [deploymentId, status, buildLogs, onConnectionChange]);

  const jumpToBottom = () => {
    terminalRef.current?.scrollToBottom();
    setIsDetached(false);
  };

  return (
    <div className="relative flex h-full w-full flex-col bg-[#000000]">
      <div 
        ref={containerRef} 
        className="flex-1 overflow-hidden p-4" 
        style={{ height: '100%' }}
      />
      {isDetached && (
        <button
          onClick={jumpToBottom}
          className="bg-accent text-accent-foreground hover:bg-accent/80 absolute right-6 bottom-6 flex items-center gap-1.5 rounded-full px-3 py-1.5 text-xs shadow-lg transition"
        >
          <ChevronDown className="h-3.5 w-3.5" /> Jump to bottom
        </button>
      )}
    </div>
  );
}
