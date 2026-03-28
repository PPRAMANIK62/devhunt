import { useEffect, useState } from "react";
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
import { Input } from "@/components/ui/input";
import { Label } from "@/components/ui/label";
import {
  Select,
  SelectContent,
  SelectItem,
  SelectTrigger,
  SelectValue,
} from "@/components/ui/select";
import { Textarea } from "@/components/ui/textarea";
import { useCreateJob, useUpdateJob } from "@/hooks/use-jobs";
import type { Job } from "@/types";

interface JobFormDrawerProps {
  open: boolean;
  onOpenChange: (open: boolean) => void;
  job?: Job | null;
  onSuccess: () => void;
}

export function JobFormDrawer({
  open,
  onOpenChange,
  job,
  onSuccess,
}: JobFormDrawerProps) {
  const isEdit = !!job;
  const { execute: createJob, loading: creating } = useCreateJob();
  const { execute: updateJob, loading: updating } = useUpdateJob();
  const loading = creating || updating;

  const [title, setTitle] = useState("");
  const [description, setDescription] = useState("");
  const [location, setLocation] = useState("");
  const [salaryMin, setSalaryMin] = useState("");
  const [salaryMax, setSalaryMax] = useState("");
  const [tags, setTags] = useState("");
  const [status, setStatus] = useState<string>("open");

  useEffect(() => {
    if (job) {
      setTitle(job.title);
      setDescription(job.description);
      setLocation(job.location);
      setSalaryMin(String(job.salary_min));
      setSalaryMax(String(job.salary_max));
      setTags(job.tags?.join(", ") ?? "");
      setStatus(job.status);
    } else {
      setTitle("");
      setDescription("");
      setLocation("");
      setSalaryMin("");
      setSalaryMax("");
      setTags("");
      setStatus("open");
    }
  }, [job, open]);

  async function handleSubmit(e: React.FormEvent<HTMLFormElement>) {
    e.preventDefault();
    const tagList = tags
      .split(",")
      .map((t) => t.trim())
      .filter(Boolean);
    try {
      if (isEdit && job) {
        await updateJob(job.id, {
          title,
          description,
          location,
          salary_min: Number(salaryMin),
          salary_max: Number(salaryMax),
          tags: tagList,
          status,
        });
        toast.success("Job updated");
      } else {
        await createJob({
          title,
          description,
          location,
          salary_min: Number(salaryMin),
          salary_max: Number(salaryMax),
          tags: tagList,
        });
        toast.success("Job posted");
      }
      onOpenChange(false);
      onSuccess();
    } catch (err) {
      toast.error(err instanceof Error ? err.message : "Something went wrong");
    }
  }

  return (
    <Drawer open={open} onOpenChange={onOpenChange}>
      <DrawerContent>
        <div className="mx-auto w-full max-w-lg overflow-y-auto">
          <DrawerHeader>
            <DrawerTitle className="font-display">
              {isEdit ? "Edit Job" : "Post a Job"}
            </DrawerTitle>
            <DrawerDescription>
              {isEdit
                ? "Update the job listing details."
                : "Fill in the details to post a new job."}
            </DrawerDescription>
          </DrawerHeader>

          <form onSubmit={handleSubmit} className="flex flex-col gap-4 px-4">
            <div className="space-y-1.5">
              <Label htmlFor="title">Job Title</Label>
              <Input
                id="title"
                value={title}
                onChange={(e) => setTitle(e.target.value)}
                placeholder="Senior Backend Engineer"
                required
              />
            </div>

            <div className="space-y-1.5">
              <Label htmlFor="description">Description</Label>
              <Textarea
                id="description"
                value={description}
                onChange={(e) => setDescription(e.target.value)}
                placeholder="Describe the role, responsibilities, and requirements..."
                rows={4}
                required
              />
            </div>

            <div className="space-y-1.5">
              <Label htmlFor="location">Location</Label>
              <Input
                id="location"
                value={location}
                onChange={(e) => setLocation(e.target.value)}
                placeholder="Remote / New York, NY"
                required
              />
            </div>

            <div className="grid grid-cols-2 gap-3">
              <div className="space-y-1.5">
                <Label htmlFor="salaryMin">Min Salary ($)</Label>
                <Input
                  id="salaryMin"
                  type="number"
                  min={0}
                  value={salaryMin}
                  onChange={(e) => setSalaryMin(e.target.value)}
                  placeholder="80000"
                  required
                />
              </div>
              <div className="space-y-1.5">
                <Label htmlFor="salaryMax">Max Salary ($)</Label>
                <Input
                  id="salaryMax"
                  type="number"
                  min={0}
                  value={salaryMax}
                  onChange={(e) => setSalaryMax(e.target.value)}
                  placeholder="120000"
                  required
                />
              </div>
            </div>

            <div className="space-y-1.5">
              <Label htmlFor="tags">Tags (comma-separated)</Label>
              <Input
                id="tags"
                value={tags}
                onChange={(e) => setTags(e.target.value)}
                placeholder="go, postgres, docker"
              />
            </div>

            {isEdit && (
              <div className="space-y-1.5">
                <Label htmlFor="status">Status</Label>
                <Select value={status} onValueChange={setStatus}>
                  <SelectTrigger id="status">
                    <SelectValue />
                  </SelectTrigger>
                  <SelectContent>
                    <SelectItem value="open">Open</SelectItem>
                    <SelectItem value="draft">Draft</SelectItem>
                    <SelectItem value="closed">Closed</SelectItem>
                  </SelectContent>
                </Select>
              </div>
            )}

            <DrawerFooter className="px-0">
              <Button type="submit" disabled={loading}>
                {loading
                  ? isEdit
                    ? "Saving..."
                    : "Posting..."
                  : isEdit
                    ? "Save Changes"
                    : "Post Job"}
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
