"use client"

import * as React from "react"
import { ChevronDownIcon } from "lucide-react"
import { cn } from "@/lib/utils"

import { Button } from "@/components/ui/button"
import { Calendar } from "@/components/ui/calendar"
import { Label } from "@/components/ui/label"
import {
  Popover,
  PopoverContent,
  PopoverTrigger,
} from "@/components/ui/popover"

interface DatePickerProps {
  value?: Date
  onChange?: (date: Date | undefined) => void
  label?: string
  placeholder?: string
  disabled?: boolean
  className?: string
  buttonClassName?: string
  id?: string
  showLabel?: boolean
  formatDate?: (date: Date) => string
}

export function DatePicker({
  value,
  onChange,
  label,
  placeholder = "Select date",
  disabled = false,
  className,
  buttonClassName,
  id,
  showLabel = true,
  formatDate = (date: Date) => date.toLocaleDateString(),
}: DatePickerProps) {
  const [open, setOpen] = React.useState(false)

  return (
    <div className={cn("flex flex-col gap-3", className)}>
      {showLabel && label && (
        <Label htmlFor={id} className="px-1">
          {label}
        </Label>
      )}
      <Popover open={open} onOpenChange={setOpen}>
        <PopoverTrigger asChild>
          <Button
            variant="outline"
            id={id}
            disabled={disabled}
            className={cn(
              "w-48 justify-between font-normal",
              !value && "text-muted-foreground",
              buttonClassName
            )}
          >
            {value ? formatDate(value) : placeholder}
            <ChevronDownIcon className="h-4 w-4" />
          </Button>
        </PopoverTrigger>
        <PopoverContent className="w-auto overflow-hidden p-0" align="start">
          <Calendar
            mode="single"
            selected={value}
            captionLayout="dropdown"
            onSelect={(date: any) => {
              onChange?.(date)
              setOpen(false)
            }}
          />
        </PopoverContent>
      </Popover>
    </div>
  )
} 