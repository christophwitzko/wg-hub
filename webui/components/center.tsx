import * as React from "react";
import { Slot } from "@radix-ui/react-slot";
import { cn } from "@/lib/utils";

export interface CenterProps extends React.HTMLAttributes<HTMLDivElement> {
  asChild?: boolean;
}

const Center = React.forwardRef<HTMLDivElement, CenterProps>(
  ({ className, asChild = false, ...props }, ref) => {
    const Comp = asChild ? Slot : "div";
    return (
      <Comp
        className={cn(
          "flex min-h-screen items-center justify-center",
          className,
        )}
        ref={ref}
        {...props}
      />
    );
  },
);
Center.displayName = "Center";

export { Center };
