/**
 * 前端时区处理工具包
 * 确保时间的正确显示和传递
 */

/**
 * 格式化时间为本地时区显示
 * @param dateString - 后端返回的UTC时间字符串
 * @param options - 格式化选项
 * @returns 本地时区格式化后的字符串
 */
export function formatToLocal(
  dateString: string | Date,
  options: Intl.DateTimeFormatOptions = {
    year: 'numeric',
    month: '2-digit',
    day: '2-digit',
    hour: '2-digit',
    minute: '2-digit',
    second: '2-digit',
  }
): string {
  if (!dateString) return '';
  
  const date = typeof dateString === 'string' ? new Date(dateString) : dateString;
  
  // 检查日期是否有效
  if (isNaN(date.getTime())) {
    console.warn('Invalid date string:', dateString);
    return '';
  }
  
  return date.toLocaleString(undefined, options);
}

/**
 * 格式化时间为本地日期显示（不含时间）
 * @param dateString - 后端返回的UTC时间字符串
 * @returns 本地时区格式化后的日期字符串
 */
export function formatDateToLocal(dateString: string | Date): string {
  return formatToLocal(dateString, {
    year: 'numeric',
    month: '2-digit',
    day: '2-digit',
  });
}

/**
 * 格式化时间为本地时间显示（不含日期）
 * @param dateString - 后端返回的UTC时间字符串
 * @returns 本地时区格式化后的时间字符串
 */
export function formatTimeToLocal(dateString: string | Date): string {
  return formatToLocal(dateString, {
    hour: '2-digit',
    minute: '2-digit',
    second: '2-digit',
  });
}

/**
 * 将本地时间转换为UTC ISO字符串，用于发送到后端
 * @param date - 本地时间Date对象
 * @returns UTC ISO字符串
 */
export function toUTCString(date: Date): string {
  if (!date || isNaN(date.getTime())) {
    console.warn('Invalid date object:', date);
    return '';
  }
  
  return date.toISOString();
}

/**
 * 解析日期字符串，保持时区信息
 * @param dateString - 日期字符串（YYYY-MM-DD格式）
 * @param timeString - 时间字符串（HH:mm格式），可选
 * @returns 带时区信息的Date对象
 */
export function parseDateTime(dateString: string, timeString?: string): Date {
  if (!dateString) {
    throw new Error('Date string is required');
  }
  
  if (timeString) {
    // 组合日期和时间，保持本地时区
    const combinedString = `${dateString}T${timeString}:00`;
    return new Date(combinedString);
  } else {
    // 只有日期，设置为当天的00:00:00本地时间
    const date = new Date(dateString + 'T00:00:00');
    return date;
  }
}

/**
 * 将日期范围转换为UTC字符串，用于API查询
 * @param startDate - 开始日期
 * @param endDate - 结束日期
 * @returns 包含start_time和end_time的对象
 */
export function formatDateRangeForAPI(startDate?: Date, endDate?: Date) {
  const result: { start_time?: string; end_time?: string } = {};
  
  if (startDate) {
    // 设置为当天开始时间（00:00:00）
    const start = new Date(startDate);
    start.setHours(0, 0, 0, 0);
    result.start_time = toUTCString(start);
  }
  
  if (endDate) {
    // 设置为当天结束时间（23:59:59）
    const end = new Date(endDate);
    end.setHours(23, 59, 59, 999);
    result.end_time = toUTCString(end);
  }
  
  return result;
}

/**
 * 获取当前用户的时区标识
 * @returns 时区标识，如 'Asia/Shanghai', 'America/New_York'
 */
export function getUserTimezone(): string {
  return Intl.DateTimeFormat().resolvedOptions().timeZone;
}

/**
 * 获取当前用户的时区偏移量
 * @returns 时区偏移量（分钟），如东8区返回-480
 */
export function getTimezoneOffset(): number {
  return new Date().getTimezoneOffset();
}

/**
 * 获取当前时间的UTC ISO字符串
 * @returns UTC ISO字符串
 */
export function nowUTC(): string {
  return new Date().toISOString();
}

/**
 * 判断两个日期是否为同一天（本地时区）
 * @param date1 - 第一个日期
 * @param date2 - 第二个日期
 * @returns 是否为同一天
 */
export function isSameDay(date1: Date, date2: Date): boolean {
  return (
    date1.getFullYear() === date2.getFullYear() &&
    date1.getMonth() === date2.getMonth() &&
    date1.getDate() === date2.getDate()
  );
}

/**
 * 创建一个显示时区信息的格式化函数
 * @param showTimezone - 是否显示时区信息
 * @returns 格式化函数
 */
export function createTimeFormatter(showTimezone: boolean = false) {
  return (dateString: string | Date) => {
    if (showTimezone) {
      return formatToLocal(dateString, {
        year: 'numeric',
        month: '2-digit',
        day: '2-digit',
        hour: '2-digit',
        minute: '2-digit',
        second: '2-digit',
        timeZoneName: 'short',
      });
    }
    return formatToLocal(dateString);
  };
}
