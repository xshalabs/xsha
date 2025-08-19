/**
 * 前端时区处理工具包
 * 确保时间的正确显示和传递
 */

/**
 * 格式化时间为本地时区显示
 * @param dateString - 后端返回的UTC时间字符串
 * @param options - 格式化选项
 * @param locale - 语言代码，如 'en-US', 'zh-CN'
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
  },
  locale?: string
): string {
  if (!dateString) return '';
  
  const date = typeof dateString === 'string' ? new Date(dateString) : dateString;
  
  // 检查日期是否有效
  if (isNaN(date.getTime())) {
    console.warn('Invalid date string:', dateString);
    return '';
  }
  
  return date.toLocaleString(locale, options);
}

/**
 * 格式化时间为本地日期显示（不含时间）
 * @param dateString - 后端返回的UTC时间字符串
 * @param locale - 语言代码
 * @returns 本地时区格式化后的日期字符串
 */
export function formatDateToLocal(dateString: string | Date, locale?: string): string {
  return formatToLocal(dateString, {
    year: 'numeric',
    month: '2-digit',
    day: '2-digit',
  }, locale);
}

/**
 * 格式化时间为本地时间显示（不含日期）
 * @param dateString - 后端返回的UTC时间字符串
 * @param locale - 语言代码
 * @returns 本地时区格式化后的时间字符串
 */
export function formatTimeToLocal(dateString: string | Date, locale?: string): string {
  return formatToLocal(dateString, {
    hour: '2-digit',
    minute: '2-digit',
    second: '2-digit',
  }, locale);
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
export function createTimeFormatter(showTimezone: boolean = false, locale?: string) {
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
      }, locale);
    }
    return formatToLocal(dateString, undefined, locale);
  };
}

/**
 * 格式化未来执行时间为人类友好的显示格式
 * @param dateString - 未来执行时间字符串
 * @param t - 国际化翻译函数
 * @param locale - 语言代码，如 'en-US', 'zh-CN'
 * @returns 人类友好的时间显示
 */
export function formatFutureExecutionTime(
  dateString: string | Date, 
  t?: (key: string, options?: any) => string,
  locale?: string
): string {
  if (!dateString) return '';
  
  const date = typeof dateString === 'string' ? new Date(dateString) : dateString;
  
  // 检查日期是否有效
  if (isNaN(date.getTime())) {
    console.warn('Invalid date string:', dateString);
    return '';
  }
  
  const now = new Date();
  const diffMs = date.getTime() - now.getTime();
  
  // 如果是过去的时间，直接返回格式化的时间
  if (diffMs <= 0) {
    return formatToLocal(date, {
      month: 'short',
      day: 'numeric',
      hour: '2-digit',
      minute: '2-digit',
    }, locale);
  }
  
  // 计算时间差
  const diffMinutes = Math.floor(diffMs / (1000 * 60));
  const diffHours = Math.floor(diffMs / (1000 * 60 * 60));
  const diffDays = Math.floor(diffMs / (1000 * 60 * 60 * 24));
  
  // 根据时间差返回不同的显示格式
  if (diffMinutes < 1) {
    // 小于1分钟，显示"即将执行"
    return t ? t('tasks.timeRelative.soon') : '即将执行';
  } else if (diffMinutes < 60) {
    return t ? t('tasks.timeRelative.minutesLater', { minutes: diffMinutes }) : `${diffMinutes}分钟后`;
  } else if (diffHours < 24) {
    return t ? t('tasks.timeRelative.hoursLater', { hours: diffHours }) : `${diffHours}小时后`;
  } else if (diffDays < 7) {
    return t ? t('tasks.timeRelative.daysLater', { days: diffDays }) : `${diffDays}天后`;
  } else {
    // 超过一周，显示具体日期时间
    return formatToLocal(date, {
      month: 'short',
      day: 'numeric',
      hour: '2-digit',
      minute: '2-digit',
    }, locale);
  }
}
