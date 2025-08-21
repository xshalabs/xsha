import { useState, useCallback, useEffect, useRef } from "react";

export function useInlineControls() {
  const [isTimePickerOpen, setIsTimePickerOpen] = useState(false);
  const [isModelSelectorOpen, setIsModelSelectorOpen] = useState(false);
  const timePickerRef = useRef<HTMLDivElement>(null);
  const modelSelectorRef = useRef<HTMLDivElement>(null);

  const handleTimePickerToggle = useCallback(() => {
    setIsTimePickerOpen(!isTimePickerOpen);
    setIsModelSelectorOpen(false); // Close model selector when opening time picker
  }, [isTimePickerOpen]);

  const handleModelSelectorToggle = useCallback(() => {
    setIsModelSelectorOpen(!isModelSelectorOpen);
    setIsTimePickerOpen(false); // Close time picker when opening model selector
  }, [isModelSelectorOpen]);

  const closeTimePickerManual = useCallback(() => {
    setIsTimePickerOpen(false);
  }, []);

  const closeModelSelectorManual = useCallback(() => {
    setIsModelSelectorOpen(false);
  }, []);

  const closeAllControls = useCallback(() => {
    setIsTimePickerOpen(false);
    setIsModelSelectorOpen(false);
  }, []);

  // Handle click outside to close popups
  useEffect(() => {
    const handleClickOutside = (event: MouseEvent) => {
      const target = event.target as Element;
      
      // Only close if clicking completely outside our components and not on any portal/popup content
      const isClickOnPortal = target.closest('[data-radix-popper-content-wrapper], [data-radix-portal], [data-sonner-toaster]');
      const isClickOnTimePicker = timePickerRef.current?.contains(target as Node);
      const isClickOnModelSelector = modelSelectorRef.current?.contains(target as Node);
      
      if (!isClickOnPortal && !isClickOnTimePicker && !isClickOnModelSelector) {
        setIsTimePickerOpen(false);
        setIsModelSelectorOpen(false);
      }
    };

    // Use a timeout to avoid immediate closure
    const timeoutId = setTimeout(() => {
      document.addEventListener('mousedown', handleClickOutside);
    }, 100);

    return () => {
      clearTimeout(timeoutId);
      document.removeEventListener('mousedown', handleClickOutside);
    };
  }, []);

  return {
    isTimePickerOpen,
    isModelSelectorOpen,
    timePickerRef,
    modelSelectorRef,
    handleTimePickerToggle,
    handleModelSelectorToggle,
    closeTimePickerManual,
    closeModelSelectorManual,
    closeAllControls,
  };
}