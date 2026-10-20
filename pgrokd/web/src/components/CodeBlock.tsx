import CopyButton from "./CopyButton";

// CodeBlock renders a copyable shell snippet. When `secret` is set and not
// revealed, it is masked on screen but still copied verbatim.
export default function CodeBlock({
  code,
  secret,
  reveal = true,
}: {
  code: string;
  secret?: string;
  reveal?: boolean;
}) {
  const shown = secret && !reveal ? code.split(secret).join("•".repeat(12)) : code;

  return (
    <div className="flex items-start gap-2 rounded border-2 border-border bg-[#1d1d1d] p-3 text-[#f5f5f5]">
      <pre className="min-w-0 flex-1 font-mono text-sm leading-6 wrap-break-word whitespace-pre-wrap">
        {shown}
      </pre>
      <CopyButton value={code} label="Copy command" className="border-[#f5f5f5] bg-[#1d1d1d] text-[#f5f5f5]" />
    </div>
  );
}
