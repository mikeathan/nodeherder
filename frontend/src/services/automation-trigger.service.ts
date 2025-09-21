import { Automation } from '@/types/automation.type';
const baseUrl = import.meta.env.VITE_API_BASE_URL;

export const triggerAutomation = async (automation: Automation, triggerName: string): Promise<boolean> => {
  const automationId = automation.id;
  const res = await fetch(`${baseUrl}/api/automation/trigger`, {
    method: 'POST',
    headers: {
      'Content-Type': 'application/json',
    },
    body: JSON.stringify({ triggerName, automationId }),
  });

  if (!res.ok) {
    console.error('Failed to trigger automation', res);
    return false;
  }
  return true;
};
