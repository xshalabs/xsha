import type { VariantProps } from "class-variance-authority";
import { cva } from "class-variance-authority";

import { cn } from "@/lib/utils";

const metricCardVariants = cva(
  "flex flex-col gap-1 border rounded-lg px-3 py-2 text-card-foreground transition-all hover:shadow-sm",
  {
    variants: {
      variant: {
        default: "border-input bg-card",
        ghost: "border-transparent",
        destructive: "border-destructive/80 bg-destructive/10",
        success: "border-success/80 bg-success/10",
        warning: "border-warning/80 bg-warning/10",
      },
    },
    defaultVariants: {
      variant: "default",
    },
  }
);

export function MetricCard({
  children,
  className,
  variant,
  ...props
}: React.ComponentProps<"div"> & VariantProps<typeof metricCardVariants>) {
  return (
    <div
      data-variant={variant}
      className={cn(metricCardVariants({ variant, className }), "group")}
      {...props}
    >
      {children}
    </div>
  );
}

export function MetricCardTitle({
  children,
  className,
  ...props
}: React.ComponentProps<"p">) {
  return (
    <p className={cn("text-sm font-medium", className)} {...props}>
      {children}
    </p>
  );
}

export function MetricCardHeader({
  children,
  className,
  ...props
}: React.ComponentProps<"div">) {
  return (
    <div
      className={cn(
        "text-muted-foreground",
        "group-data-[variant=destructive]:text-destructive",
        "group-data-[variant=success]:text-success",
        "group-data-[variant=warning]:text-warning",
        className
      )}
      {...props}
    >
      {children}
    </div>
  );
}

export function MetricCardValue({
  children,
  className,
  ...props
}: React.ComponentProps<"p">) {
  return (
    <p className={cn("text-foreground font-semibold", className)} {...props}>
      {children}
    </p>
  );
}

export function MetricCardGroup({
  children,
  className,
  ...props
}: React.ComponentProps<"div">) {
  return (
    <div
      className={cn(
        "grid grid-cols-2 sm:grid-cols-3 md:grid-cols-4 lg:grid-cols-5 gap-4",
        className
      )}
      {...props}
    >
      {children}
    </div>
  );
}

const metricCardButtonVariants = cva(
  "group w-full text-left transition-all rounded-md outline-none focus-visible:border-ring focus-visible:ring-ring/50 focus-visible:ring-[3px] cursor-pointer hover:shadow-md"
);

export function MetricCardButton({
  children,
  className,
  variant,
  ...props
}: React.ComponentProps<"button"> & VariantProps<typeof metricCardVariants>) {
  return (
    <button
      type="button"
      data-variant={variant}
      className={cn(
        metricCardVariants({ variant, className }),
        metricCardButtonVariants()
      )}
      {...props}
    >
      {children}
    </button>
  );
}
