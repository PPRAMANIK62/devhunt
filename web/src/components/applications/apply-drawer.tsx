import { useState } from "react";
import { toast } from "sonner";
import { Button } from "@/components/ui/button";
import {
  Drawer,
  DrawerContent,
  DrawerDescription,
  DrawerHeader,
  DrawerTitle,
} from "@/components/ui/drawer";
import { Label } from "@/components/ui/label";
import { Textarea } from "@/components/ui/textarea";
import { useApply } from "@/hooks/use-applications";

interface ApplyDrawerProps {
  open: boolean;
  onOpenChange: (open: boolean) => void;
  jobId: string;
  jobTitle: string;
  onSuccess?: () => void;
}

export function ApplyDrawer({
  open,
  onOpenChange,
  jobId,
  jobTitle,
  onSuccess,
}: ApplyDrawerProps) {
  const [coverNote, setCoverNote] = useState("");
  const { execute: apply, loading } = useApply();

  async function handleSubmit(e: React.FormEvent<HTMLFormElement>) {
    e.preventDefault();
    try {
      await apply(jobId, coverNote);
      toast.success("Application submitted!");
      setCoverNote("");
      onOpenChange(false);
      onSuccess?.();
    } catch (err) {
      toast.error(err instanceof Error ? err.message : "Failed to apply");
    }
  }

  return (
    <Drawer open={open} onOpenChange={onOpenChange} direction="right" shouldScaleBackground={false}>
      <DrawerContent className="inset-y-0 right-0 left-auto mt-0 h-full w-full max-w-2xl rounded-none p-0 flex flex-col [&>div:first-child]:hidden">
        <DrawerHeader className="px-6 py-4 border-b border-border shrink-0">
          <DrawerTitle className="font-display">Apply</DrawerTitle>
          <DrawerDescription className="line-clamp-2 text-sm text-muted-foreground">
            {jobTitle}
          </DrawerDescription>
        </DrawerHeader>

        <form onSubmit={handleSubmit} className="flex flex-1 flex-col overflow-y-auto">
          <div className="flex flex-col gap-4 px-6 py-5">
            <div className="space-y-1.5">
              <Label htmlFor="coverNote">
                Cover Note{" "}
                <span className="text-muted-foreground">(optional)</span>
              </Label>
              <Textarea
                id="coverNote"
                value={coverNote}
                onChange={(e) => setCoverNote(e.target.value)}
                placeholder="Tell the company why you're a great fit..."
                rows={5}
                maxLength={2000}
              />
              <p className="text-right font-mono text-xs text-muted-foreground">
                {coverNote.length}/2000
              </p>
            </div>
          </div>

          <div className="mt-auto border-t border-border px-6 py-4 flex flex-col gap-2">
            <Button type="submit" disabled={loading}>
              {loading ? "Submitting..." : "Submit Application"}
            </Button>
            <Button type="button" variant="outline" onClick={() => onOpenChange(false)}>
              Cancel
            </Button>
          </div>
        </form>
      </DrawerContent>
    </Drawer>
  );
}
