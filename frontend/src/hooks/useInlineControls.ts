import { useState, useCallback, useEffect, useRef } from "react";

export function useInlineControls() {
  const [isTimePickerOpen, setIsTimePickerOpen] = useState(false);
  const [isModelSelectorOpen, setIsModelSelectorOpen] = useState(false);
  const [isPlanModeSelectorOpen, setIsPlanModeSelectorOpen] = useState(false);
  const timePickerRef = useRef<HTMLDivElement>(null);
  const modelSelectorRef = useRef<HTMLDivElement>(null);
  const planModeSelectorRef = useRef<HTMLDivElement>(null);

  const handleTimePickerToggle = useCallback(() => {
    setIsTimePickerOpen(!isTimePickerOpen);
    setIsModelSelectorOpen(false); // Close model selector when opening time picker
    setIsPlanModeSelectorOpen(false); // Close plan mode selector when opening time picker
  }, [isTimePickerOpen]);

  const handleModelSelectorToggle = useCallback(() => {
    setIsModelSelectorOpen(!isModelSelectorOpen);
    setIsTimePickerOpen(false); // Close time picker when opening model selector
    setIsPlanModeSelectorOpen(false); // Close plan mode selector when opening model selector
  }, [isModelSelectorOpen]);

  const handlePlanModeSelectorToggle = useCallback(() => {
    setIsPlanModeSelectorOpen(!isPlanModeSelectorOpen);
    setIsTimePickerOpen(false); // Close time picker when opening plan mode selector
    setIsModelSelectorOpen(false); // Close model selector when opening plan mode selector
  }, [isPlanModeSelectorOpen]);

  const closeTimePickerManual = useCallback(() => {
    setIsTimePickerOpen(false);
  }, []);

  const closeModelSelectorManual = useCallback(() => {
    setIsModelSelectorOpen(false);
  }, []);

  const closePlanModeSelectorManual = useCallback(() => {
    setIsPlanModeSelectorOpen(false);
  }, []);

  const closeAllControls = useCallback(() => {
    setIsTimePickerOpen(false);
    setIsModelSelectorOpen(false);
    setIsPlanModeSelectorOpen(false);
  }, []);

  // Handle click outside to close popups
  useEffect(() => {
    const handleClickOutside = (event: MouseEvent) => {
      const target = event.target as Element;
      
      // Only close if clicking completely outside our components and not on any portal/popup content
      const isClickOnPortal = target.closest('[data-radix-popper-content-wrapper], [data-radix-portal], [data-sonner-toaster]');
      const isClickOnTimePicker = timePickerRef.current?.contains(target as Node);
      const isClickOnModelSelector = modelSelectorRef.current?.contains(target as Node);
      const isClickOnPlanModeSelector = planModeSelectorRef.current?.contains(target as Node);
      
      if (!isClickOnPortal && !isClickOnTimePicker && !isClickOnModelSelector && !isClickOnPlanModeSelector) {
        setIsTimePickerOpen(false);
        setIsModelSelectorOpen(false);
        setIsPlanModeSelectorOpen(false);
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
    isPlanModeSelectorOpen,
    timePickerRef,
    modelSelectorRef,
    planModeSelectorRef,
    handleTimePickerToggle,
    handleModelSelectorToggle,
    handlePlanModeSelectorToggle,
    closeTimePickerManual,
    closeModelSelectorManual,
    closePlanModeSelectorManual,
    closeAllControls,
  };
}