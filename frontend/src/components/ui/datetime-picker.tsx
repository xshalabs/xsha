"use client"

import * as React from "react"
import { CalendarIcon, ClockIcon } from "lucide-react"
import { cn } from "@/lib/utils"
import { format } from "date-fns"

import { Button } from "@/components/ui/button"
import { Calendar } from "@/components/ui/calendar"
import { Input } from "@/components/ui/input"
import { Label } from "@/components/ui/label"
import {
  Popover,
  PopoverContent,
  PopoverTrigger,
} from "@/components/ui/popover"

interface DateTimePickerProps {
  value?: Date
  onChange?: (date: Date | undefined) => void
  label?: string
  placeholder?: string
  disabled?: boolean
  className?: string
  id?: string
  error?: string
}

export function DateTimePicker({
  value,
  onChange,
  label,
  placeholder = "选择执行时间",
  disabled = false,
  className,
  id,
  error,
}: DateTimePickerProps) {
  const [open, setOpen] = React.useState(false)
  const [timeValue, setTimeValue] = React.useState("")

  // 当 value 变化时，更新时间输入框的值
  React.useEffect(() => {
    if (value) {
      const hours = value.getHours().toString().padStart(2, '0')
      const minutes = value.getMinutes().toString().padStart(2, '0')
      setTimeValue(`${hours}:${minutes}`)
    } else {
      setTimeValue("")
    }
  }, [value])

  const handleDateSelect = (selectedDate: Date | undefined) => {
    if (!selectedDate) {
      onChange?.(undefined)
      return
    }

    let newDate = new Date(selectedDate)
    
    // 如果已有时间值，保持时间部分
    if (timeValue && timeValue.includes(':')) {
      const [hours, minutes] = timeValue.split(':').map(Number)
      if (!isNaN(hours) && !isNaN(minutes)) {
        newDate.setHours(hours, minutes, 0, 0)
      }
    } else {
      // 默认设置为当前时间
      const now = new Date()
      newDate.setHours(now.getHours(), now.getMinutes(), 0, 0)
    }

    onChange?.(newDate)
  }

  const handleTimeChange = (e: React.ChangeEvent<HTMLInputElement>) => {
    const newTimeValue = e.target.value
    setTimeValue(newTimeValue)

    if (value && newTimeValue.includes(':')) {
      const [hours, minutes] = newTimeValue.split(':').map(Number)
      if (!isNaN(hours) && !isNaN(minutes) && hours >= 0 && hours <= 23 && minutes >= 0 && minutes <= 59) {
        const newDate = new Date(value)
        newDate.setHours(hours, minutes, 0, 0)
        onChange?.(newDate)
      }
    }
  }

  const formatDateTime = (date: Date) => {
    return format(date, "yyyy-MM-dd HH:mm")
  }

  const clearSelection = () => {
    onChange?.(undefined)
    setTimeValue("")
    setOpen(false)
  }

  return (
    <div className={cn("flex flex-col gap-3", className)}>
      {label && (
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
              "w-full justify-start text-left font-normal",
              !value && "text-muted-foreground",
              error && "border-red-500"
            )}
          >
            <CalendarIcon className="mr-2 h-4 w-4" />
            {value ? formatDateTime(value) : placeholder}
          </Button>
        </PopoverTrigger>
        <PopoverContent className="w-auto p-0" align="start">
          <div className="flex">
            <Calendar
              mode="single"
              selected={value}
              onSelect={handleDateSelect}
              initialFocus
              className="rounded-md"
            />
            <div className="border-l border-border p-3 space-y-3">
              <div className="space-y-2">
                <Label htmlFor="time-input" className="text-sm font-medium">
                  时间
                </Label>
                <div className="flex items-center space-x-2">
                  <ClockIcon className="h-4 w-4 text-muted-foreground" />
                  <Input
                    id="time-input"
                    type="time"
                    value={timeValue}
                    onChange={handleTimeChange}
                    className="w-32"
                  />
                </div>
              </div>
              <div className="flex space-x-2">
                <Button
                  size="sm"
                  onClick={() => setOpen(false)}
                  className="flex-1"
                >
                  确定
                </Button>
                <Button
                  size="sm"
                  variant="outline"
                  onClick={clearSelection}
                  className="flex-1"
                >
                  清除
                </Button>
              </div>
            </div>
          </div>
        </PopoverContent>
      </Popover>
      {error && (
        <p className="text-sm text-red-500">{error}</p>
      )}
    </div>
  )
}
