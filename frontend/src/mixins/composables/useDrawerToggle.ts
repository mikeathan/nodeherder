export function useDrawerToggle(isExpanded: Ref<boolean>, emitToggle: () => void) {
  const toggleDrawer = () => {
    emitToggle();
  };

  const closeDrawer = () => {
    if (isExpanded.value) emitToggle();
  };

  return { toggleDrawer, closeDrawer };
}