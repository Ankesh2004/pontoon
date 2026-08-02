import { X } from 'lucide-react';

interface ConfirmModalProps {
  title: string;
  description: string;
  confirmText?: string;
  cancelText?: string;
  onConfirm: () => void;
  onCancel: () => void;
  isLoading?: boolean;
  isDestructive?: boolean;
}

export function ConfirmModal({
  title,
  description,
  confirmText = 'Confirm',
  cancelText = 'Cancel',
  onConfirm,
  onCancel,
  isLoading = false,
  isDestructive = false,
}: ConfirmModalProps) {
  return (
    <div className="fixed inset-0 z-[100] flex items-center justify-center bg-black/50 backdrop-blur-sm">
      <div className="border-border bg-card w-full max-w-md rounded-lg border p-6 shadow-xl">
        <div className="mb-4 flex items-center justify-between">
          <h2 className="text-foreground text-xl font-bold">{title}</h2>
          <button
            onClick={onCancel}
            disabled={isLoading}
            className="text-muted-foreground hover:text-foreground disabled:opacity-50"
          >
            <X className="h-5 w-5" />
          </button>
        </div>

        <p className="text-muted-foreground mb-6 text-sm leading-relaxed">
          {description}
        </p>

        <div className="flex gap-3 pt-2">
          <button
            onClick={onCancel}
            disabled={isLoading}
            className="border-border text-foreground hover:bg-accent flex-1 rounded-md border px-4 py-2 text-sm font-medium disabled:opacity-50"
          >
            {cancelText}
          </button>
          <button
            onClick={onConfirm}
            disabled={isLoading}
            className={`flex-1 rounded-md px-4 py-2 text-sm font-medium disabled:opacity-50 ${
              isDestructive
                ? 'bg-destructive text-destructive-foreground hover:bg-destructive/90'
                : 'bg-primary text-primary-foreground hover:bg-primary/90'
            }`}
          >
            {isLoading ? 'Loading...' : confirmText}
          </button>
        </div>
      </div>
    </div>
  );
}
