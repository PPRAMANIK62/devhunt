import { ChevronDown, Search, X } from "lucide-react";
import { useEffect, useRef, useState } from "react";
import { Button } from "@/components/ui/button";
import {
  DropdownMenu,
  DropdownMenuCheckboxItem,
  DropdownMenuContent,
  DropdownMenuRadioGroup,
  DropdownMenuRadioItem,
  DropdownMenuTrigger,
} from "@/components/ui/dropdown-menu";
import { Input } from "@/components/ui/input";
import { useJobFilterOptions, type JobFilters } from "@/hooks/use-jobs";

const SALARY_OPTIONS = [
  { label: "Any salary", value: 0 },
  { label: "$50k+", value: 50000 },
  { label: "$80k+", value: 80000 },
  { label: "$100k+", value: 100000 },
  { label: "$130k+", value: 130000 },
  { label: "$150k+", value: 150000 },
  { label: "$200k+", value: 200000 },
];

interface JobFiltersProps {
  filters: JobFilters;
  onChange: (filters: JobFilters) => void;
}

function MultiSelectDropdown({
  label,
  options,
  selected,
  onToggle,
}: {
  label: string;
  options: string[];
  selected: string[];
  onToggle: (value: string) => void;
}) {
  const count = selected.length;
  return (
    <DropdownMenu>
      <DropdownMenuTrigger asChild>
        <Button
          variant="outline"
          size="sm"
          className="h-9 w-full justify-between px-3 font-mono text-sm font-normal"
        >
          <span className={count > 0 ? "text-foreground" : "text-muted-foreground"}>
            {count > 0 ? `${label} · ${count}` : label}
          </span>
          <ChevronDown className="h-3.5 w-3.5 shrink-0 text-muted-foreground" />
        </Button>
      </DropdownMenuTrigger>
      <DropdownMenuContent className="max-h-64 w-52 overflow-y-auto" align="start">
        {options.map((opt) => (
          <DropdownMenuCheckboxItem
            key={opt}
            checked={selected.includes(opt)}
            onCheckedChange={() => onToggle(opt)}
            className="font-mono text-sm"
          >
            {opt}
          </DropdownMenuCheckboxItem>
        ))}
      </DropdownMenuContent>
    </DropdownMenu>
  );
}

export function JobFiltersSidebar({ filters, onChange }: JobFiltersProps) {
  const [searchInput, setSearchInput] = useState(filters.search);
  const debounceRef = useRef<ReturnType<typeof setTimeout> | null>(null);
  const { data: options } = useJobFilterOptions();

  const isActive =
    filters.search || filters.locations.length > 0 || filters.tags.length > 0 || filters.minSalary > 0;

  const salaryLabel =
    SALARY_OPTIONS.find((o) => o.value === filters.minSalary)?.label ?? "Any salary";

  useEffect(() => {
    setSearchInput(filters.search);
  }, [filters.search]);

  useEffect(() => {
    return () => {
      if (debounceRef.current) clearTimeout(debounceRef.current);
    };
  }, []);

  function handleSearchChange(value: string) {
    setSearchInput(value);
    if (debounceRef.current) clearTimeout(debounceRef.current);
    debounceRef.current = setTimeout(() => {
      onChange({ ...filters, search: value });
    }, 300);
  }

  function toggleLocation(loc: string) {
    const next = filters.locations.includes(loc)
      ? filters.locations.filter((l) => l !== loc)
      : [...filters.locations, loc];
    onChange({ ...filters, locations: next });
  }

  function toggleTag(tag: string) {
    const next = filters.tags.includes(tag)
      ? filters.tags.filter((t) => t !== tag)
      : [...filters.tags, tag];
    onChange({ ...filters, tags: next });
  }

  function handleClear() {
    setSearchInput("");
    onChange({ search: "", locations: [], tags: [], minSalary: 0 });
  }

  return (
    <aside className="w-full space-y-2 lg:w-52 lg:shrink-0">
      <div className="relative">
        <Search className="absolute top-1/2 left-3 h-3.5 w-3.5 -translate-y-1/2 text-muted-foreground" />
        <Input
          value={searchInput}
          onChange={(e) => handleSearchChange(e.target.value)}
          placeholder="Search roles…"
          className="h-9 pl-8 font-mono text-sm"
        />
      </div>

      <MultiSelectDropdown
        label="Location"
        options={options.locations}
        selected={filters.locations}
        onToggle={toggleLocation}
      />

      <MultiSelectDropdown
        label="Tags"
        options={options.tags}
        selected={filters.tags}
        onToggle={toggleTag}
      />

      <DropdownMenu>
        <DropdownMenuTrigger asChild>
          <Button
            variant="outline"
            size="sm"
            className="h-9 w-full justify-between px-3 font-mono text-sm font-normal"
          >
            <span className={filters.minSalary > 0 ? "text-foreground" : "text-muted-foreground"}>
              {salaryLabel}
            </span>
            <ChevronDown className="h-3.5 w-3.5 shrink-0 text-muted-foreground" />
          </Button>
        </DropdownMenuTrigger>
        <DropdownMenuContent className="w-44" align="start">
          <DropdownMenuRadioGroup
            value={String(filters.minSalary)}
            onValueChange={(v) => onChange({ ...filters, minSalary: Number(v) })}
          >
            {SALARY_OPTIONS.map((o) => (
              <DropdownMenuRadioItem key={o.value} value={String(o.value)} className="font-mono text-sm">
                {o.label}
              </DropdownMenuRadioItem>
            ))}
          </DropdownMenuRadioGroup>
        </DropdownMenuContent>
      </DropdownMenu>

      {isActive && (
        <button
          onClick={handleClear}
          className="flex items-center gap-1 pt-1 font-mono text-xs text-muted-foreground hover:text-foreground"
        >
          <X className="h-3 w-3" />
          Clear filters
        </button>
      )}
    </aside>
  );
}
