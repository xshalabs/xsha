import {
  Breadcrumb,
  BreadcrumbItem,
  BreadcrumbLink,
  BreadcrumbList,
  BreadcrumbPage,
  BreadcrumbSeparator,
} from "@/components/ui/breadcrumb";
import {
  Select,
  SelectContent,
  SelectItem,
  SelectTrigger,
  SelectValue,
} from "@/components/ui/select";
import { type LucideIcon } from "lucide-react";
import { Fragment } from "react";
import { Link } from "react-router-dom";

interface NavBreadcrumbProps {
  items: (
    | {
        type: "link";
        label: string;
        href: string;
      }
    | {
        type: "page";
        label: string;
      }
    | {
        type: "select";
        value: string;
        items: { value: string; label: string; icon?: LucideIcon }[];
        onValueChange: (value: string) => void;
      }
  )[];
}

export function NavBreadcrumb({ items }: NavBreadcrumbProps) {
  return (
    <Breadcrumb>
      <BreadcrumbList>
        {items.map((item, i) => (
          <Fragment key={i}>
            <BreadcrumbItem>
              {item.type === "link" ? (
                <BreadcrumbLink className="hidden md:block" asChild>
                  <Link to={item.href}>{item.label}</Link>
                </BreadcrumbLink>
              ) : null}
              {item.type === "page" ? (
                <BreadcrumbPage className="hidden md:block max-w-[120px] truncate lg:max-w-[200px]">
                  {item.label}
                </BreadcrumbPage>
              ) : null}
              {item.type === "select" ? (
                <Select
                  value={item.value}
                  onValueChange={item.onValueChange}
                >
                  <SelectTrigger
                    className="text-foreground [&>span_svg]:text-muted-foreground/80 [&>span]:flex [&>span]:items-center [&>span]:gap-2 [&>span_svg]:shrink-0 border-none shadow-none bg-transparent min-w-48"
                    aria-label="Select option"
                  >
                    <SelectValue />
                  </SelectTrigger>
                  <SelectContent>
                    {item.items.map((selectItem, j) => (
                      <SelectItem key={j} value={selectItem.value}>
                        {selectItem.icon && (
                          <selectItem.icon size={16} aria-hidden="true" className="mr-2" />
                        )}
                        {selectItem.label}
                      </SelectItem>
                    ))}
                  </SelectContent>
                </Select>
              ) : null}
            </BreadcrumbItem>
            {i < items.length - 1 && (
              <BreadcrumbSeparator className="hidden md:block" />
            )}
          </Fragment>
        ))}
      </BreadcrumbList>
    </Breadcrumb>
  );
}
