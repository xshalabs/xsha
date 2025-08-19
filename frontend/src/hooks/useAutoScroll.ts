import { useState, useCallback } from 'react';

export interface UseAutoScrollReturn {
  autoScroll: boolean;
  toggleAutoScroll: () => void;
  setAutoScroll: (value: boolean) => void;
}

export const useAutoScroll = (initialValue = true): UseAutoScrollReturn => {
  const [autoScroll, setAutoScroll] = useState(initialValue);

  const toggleAutoScroll = useCallback(() => {
    setAutoScroll(prev => !prev);
  }, []);

  return {
    autoScroll,
    toggleAutoScroll,
    setAutoScroll,
  };
};
