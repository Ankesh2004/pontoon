import { useEffect, useRef } from 'react';

interface LogTerminalProps {
  logs: string[];
  autoScroll: boolean;
}

export function LogTerminal({ logs, autoScroll }: LogTerminalProps) {
  const endRef = useRef<HTMLDivElement | null>(null);

  useEffect(() => {
    if (autoScroll) {
      endRef.current?.scrollIntoView({ block: 'end' });
    }
  }, [autoScroll, logs]);

  return (
    <div className="h-full overflow-auto rounded-md bg-black p-4 font-mono text-sm text-green-400">
      {logs.length === 0 ? (
        <div className="text-muted-foreground">Waiting for deployment logs...</div>
      ) : (
        logs.map((line, index) => <div key={`${index}-${line}`}>{line}</div>)
      )}
      <div ref={endRef} />
    </div>
  );
}
