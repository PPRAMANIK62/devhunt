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
import { Textarea } from "@/components/ui/textarea";
import { useCreateCompany, useUpdateCompany } from "@/hooks/use-company";
import type { Company } from "@/types";

interface CompanyFormDrawerProps {
  open: boolean;
  onOpenChange: (open: boolean) => void;
  company?: Company | null;
  onSuccess: (company: Company) => void;
}

export function CompanyFormDrawer({
  open,
  onOpenChange,
  company,
  onSuccess,
}: CompanyFormDrawerProps) {
  const isEdit = !!company;
  const { execute: createCompany, loading: creating } = useCreateCompany();
  const { execute: updateCompany, loading: updating } = useUpdateCompany();
  const loading = creating || updating;

  const [name, setName] = useState("");
  const [slug, setSlug] = useState("");
  const [description, setDescription] = useState("");
  const [website, setWebsite] = useState("");

  useEffect(() => {
    if (company) {
      setName(company.name);
      setSlug(company.slug);
      setDescription(company.description ?? "");
      setWebsite(company.website ?? "");
    } else {
      setName("");
      setSlug("");
      setDescription("");
      setWebsite("");
    }
  }, [company, open]);

  function handleNameChange(val: string) {
    setName(val);
    if (!isEdit) {
      setSlug(
        val
          .toLowerCase()
          .replace(/[^a-z0-9]+/g, "-")
          .replace(/^-|-$/g, ""),
      );
    }
  }

  async function handleSubmit(e: React.FormEvent<HTMLFormElement>) {
    e.preventDefault();
    try {
      let result: Company;
      if (isEdit) {
        result = await updateCompany({ name, slug, description, website });
        toast.success("Company profile updated");
      } else {
        result = await createCompany({ name, slug, description, website });
        toast.success("Company profile created");
      }
      onOpenChange(false);
      onSuccess(result);
    } catch (err) {
      toast.error(err instanceof Error ? err.message : "Something went wrong");
    }
  }

  return (
    <Drawer open={open} onOpenChange={onOpenChange}>
      <DrawerContent>
        <div className="mx-auto w-full max-w-lg">
          <DrawerHeader>
            <DrawerTitle className="font-display">
              {isEdit ? "Edit Company" : "Create Company Profile"}
            </DrawerTitle>
            <DrawerDescription>
              {isEdit
                ? "Update your company details."
                : "Set up your company profile to start posting jobs."}
            </DrawerDescription>
          </DrawerHeader>

          <form onSubmit={handleSubmit} className="flex flex-col gap-4 px-4">
            <div className="space-y-1.5">
              <Label htmlFor="name">Company Name</Label>
              <Input
                id="name"
                value={name}
                onChange={(e) => handleNameChange(e.target.value)}
                placeholder="Acme Corp"
                required
              />
            </div>

            <div className="space-y-1.5">
              <Label htmlFor="slug">Slug</Label>
              <Input
                id="slug"
                value={slug}
                onChange={(e) => setSlug(e.target.value)}
                placeholder="acme-corp"
                required
              />
              <p className="font-mono text-xs text-muted-foreground">
                URL-friendly identifier (lowercase, hyphens only)
              </p>
            </div>

            <div className="space-y-1.5">
              <Label htmlFor="desc">
                Description{" "}
                <span className="text-muted-foreground">(optional)</span>
              </Label>
              <Textarea
                id="desc"
                value={description}
                onChange={(e) => setDescription(e.target.value)}
                placeholder="What does your company do?"
                rows={3}
              />
            </div>

            <div className="space-y-1.5">
              <Label htmlFor="website">
                Website{" "}
                <span className="text-muted-foreground">(optional)</span>
              </Label>
              <Input
                id="website"
                type="url"
                value={website}
                onChange={(e) => setWebsite(e.target.value)}
                placeholder="https://acme.com"
              />
            </div>

            <DrawerFooter className="px-0">
              <Button type="submit" disabled={loading}>
                {loading
                  ? isEdit
                    ? "Saving..."
                    : "Creating..."
                  : isEdit
                    ? "Save Changes"
                    : "Create Profile"}
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
