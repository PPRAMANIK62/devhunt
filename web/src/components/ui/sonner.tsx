import {
  CircleCheck,
  Info,
  LoaderCircle,
  OctagonX,
  TriangleAlert,
} from "lucide-react";
import { Toaster as Sonner } from "sonner";

type ToasterProps = React.ComponentProps<typeof Sonner>;

const Toaster = ({ ...props }: ToasterProps) => {
  return (
    <Sonner
      theme="light"
      position="bottom-right"
      className="toaster group"
      icons={{
        success: <CircleCheck className="h-4 w-4 text-green-600" />,
        info: <Info className="h-4 w-4 text-blue-600" />,
        warning: <TriangleAlert className="h-4 w-4 text-amber-600" />,
        error: <OctagonX className="h-4 w-4 text-red-600" />,
        loading: <LoaderCircle className="h-4 w-4 animate-spin text-muted-foreground" />,
      }}
      toastOptions={{
        classNames: {
          toast: [
            "group toast font-sans text-sm",
            "bg-card text-card-foreground",
            "border border-border",
            "shadow-sm rounded-md",
            "px-4 py-3",
          ].join(" "),
          title: "font-medium text-foreground",
          description: "text-muted-foreground font-mono text-xs mt-0.5",
          actionButton: "bg-primary text-primary-foreground text-xs font-medium",
          cancelButton: "bg-muted text-muted-foreground text-xs",
          success: "border-l-4 border-l-green-500",
          error: "border-l-4 border-l-red-500",
          warning: "border-l-4 border-l-amber-500",
          info: "border-l-4 border-l-blue-500",
        },
      }}
      {...props}
    />
  );
};

export { Toaster };
