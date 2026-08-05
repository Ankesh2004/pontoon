import { CheckCircle, XCircle, AlertTriangle, ShieldCheck, ShieldAlert, Code2 } from 'lucide-react';
import { useState } from 'react';

interface ReviewCardProps {
  pipelineId: string;
  context: {
    root_cause?: string;
    proposed_patch?: string;
    security_passed?: boolean;
    confidence_score?: number;
  };
  onApprove: () => void;
  onReject: () => void;
}

export function ReviewCard({ context, onApprove, onReject }: ReviewCardProps) {
  const [isSubmitting, setIsSubmitting] = useState(false);

  const handleApprove = async () => {
    setIsSubmitting(true);
    await onApprove();
    setIsSubmitting(false);
  };

  const handleReject = async () => {
    setIsSubmitting(true);
    await onReject();
    setIsSubmitting(false);
  };

  return (
    <div className="rounded-xl border border-gray-700 bg-gray-800 shadow-xl overflow-hidden mt-6 mb-6">
      <div className="bg-gray-900 px-6 py-4 border-b border-gray-700 flex items-center justify-between">
        <div className="flex items-center gap-3">
          <div className="bg-indigo-500/20 p-2 rounded-lg">
            <Code2 className="w-5 h-5 text-indigo-400" />
          </div>
          <h3 className="text-lg font-semibold text-white">✨ Pontoon AI: Build Recovery</h3>
        </div>
        {context.confidence_score !== undefined && (
          <div className={`px-3 py-1 rounded-full text-xs font-medium flex items-center gap-1
            ${context.confidence_score >= 80 ? 'bg-green-500/20 text-green-400' : 'bg-yellow-500/20 text-yellow-400'}`}>
            Confidence: {context.confidence_score}%
          </div>
        )}
      </div>

      <div className="p-6 space-y-6 text-gray-300">
        {/* Root Cause Section */}
        <div className="space-y-2">
          <h4 className="text-sm font-medium text-gray-400 uppercase tracking-wider flex items-center gap-2">
            <AlertTriangle className="w-4 h-4 text-orange-400" />
            Root Cause Analysis
          </h4>
          <div className="bg-gray-900/50 p-4 rounded-lg border border-gray-700/50 text-sm">
            {context.root_cause || "Analyzing..."}
          </div>
        </div>

        {/* Security Section */}
        <div className="space-y-2">
          <h4 className="text-sm font-medium text-gray-400 uppercase tracking-wider flex items-center gap-2">
            {context.security_passed ? (
              <ShieldCheck className="w-4 h-4 text-emerald-400" />
            ) : (
              <ShieldAlert className="w-4 h-4 text-red-400" />
            )}
            SecOps Verification
          </h4>
          <div className={`p-4 rounded-lg border text-sm flex items-start gap-3
            ${context.security_passed 
              ? 'bg-emerald-500/10 border-emerald-500/20 text-emerald-200' 
              : 'bg-red-500/10 border-red-500/20 text-red-200'}`}>
            <span>
              {context.security_passed 
                ? "Passed (No elevated privileges or suspicious dependencies found)." 
                : "Failed (Security risks detected in the proposed patch)."}
            </span>
          </div>
        </div>

        {/* Patch Section */}
        <div className="space-y-2">
          <h4 className="text-sm font-medium text-gray-400 uppercase tracking-wider flex items-center gap-2">
            <Code2 className="w-4 h-4 text-blue-400" />
            Proposed Fix
          </h4>
          <pre className="bg-gray-950 p-4 rounded-lg border border-gray-700/50 text-xs text-gray-300 overflow-x-auto whitespace-pre-wrap font-mono">
            {context.proposed_patch || "No patch generated."}
          </pre>
        </div>
      </div>

      {/* Actions */}
      <div className="bg-gray-900/50 px-6 py-4 border-t border-gray-700 flex items-center justify-end gap-3">
        <button
          onClick={handleReject}
          disabled={isSubmitting}
          className="px-4 py-2 rounded-lg text-sm font-medium text-gray-300 hover:text-white hover:bg-gray-700 transition-colors flex items-center gap-2 disabled:opacity-50"
        >
          <XCircle className="w-4 h-4" />
          Reject
        </button>
        <button
          onClick={handleApprove}
          disabled={isSubmitting || !context.security_passed}
          className="px-4 py-2 rounded-lg text-sm font-medium bg-indigo-500 text-white hover:bg-indigo-600 transition-colors shadow-lg shadow-indigo-500/20 flex items-center gap-2 disabled:opacity-50 disabled:cursor-not-allowed"
        >
          <CheckCircle className="w-4 h-4" />
          Approve & Redeploy
        </button>
      </div>
    </div>
  );
}
