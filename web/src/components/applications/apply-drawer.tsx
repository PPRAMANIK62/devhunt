import { useState } from "react";
import { toast } from "sonner";
import { Button } from "@/components/ui/button";
import {
  Drawer,
  DrawerContent,
  DrawerDescription,
  DrawerFooter,
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
    <Drawer open={open} onOpenChange={onOpenChange}>
      <DrawerContent>
        <div className="mx-auto w-full max-w-lg">
          <DrawerHeader>
            <DrawerTitle className="font-display">Apply</DrawerTitle>
            <DrawerDescription className="line-clamp-2">
              {jobTitle}
            </DrawerDescription>
          </DrawerHeader>

          <form
            onSubmit={handleSubmit}
            className="flex flex-col gap-4 px-4"
          >
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

            <DrawerFooter className="px-0">
              <Button type="submit" disabled={loading}>
                {loading ? "Submitting..." : "Submit Application"}
              </Button>
              <Button
                type="button"
                variant="outline"
                onClick={() => onOpenChange(false)}
              >
                Cancel
              </Button>
            </DrawerFooter>
          </form>
        </div>
      </DrawerContent>
    </Drawer>
  );
}
