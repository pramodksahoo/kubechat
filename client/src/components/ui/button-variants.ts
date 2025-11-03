import { cva } from "class-variance-authority";

export const buttonVariants = cva(
  "inline-flex items-center justify-center gap-2 whitespace-nowrap rounded-lg text-sm font-semibold tracking-wide ring-offset-background transition-all focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-ring focus-visible:ring-offset-2 disabled:pointer-events-none disabled:opacity-60",
  {
    variants: {
      variant: {
        default: "bg-primary text-primary-foreground shadow-raised hover:bg-primary/85 hover:shadow-feature active:scale-[0.99]",
        destructive:
          "bg-destructive text-destructive-foreground shadow-raised hover:bg-destructive/85",
        outline:
          "border border-border bg-transparent text-foreground shadow-subtle hover:bg-accent/40",
        secondary: "bg-secondary text-secondary-foreground shadow-subtle hover:bg-secondary/80",
        ghost: "text-foreground hover:bg-accent/40 hover:text-foreground",
        link: "text-primary underline-offset-4 hover:text-primary/80 hover:underline",
      },
      size: {
        default: "h-10 px-4",
        sm: "h-9 rounded-md px-3",
        lg: "h-11 rounded-lg px-8",
        icon: "h-10 w-10 rounded-lg",
      },
    },
    defaultVariants: {
      variant: "default",
      size: "default",
    },
  },
);

export type ButtonVariants = typeof buttonVariants;
