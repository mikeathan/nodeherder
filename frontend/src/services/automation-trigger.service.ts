import { Automation } from '@/types/automation.type';
const baseUrl = import.meta.env.VITE_API_BASE_URL;

export const triggerAutomation = async (automation: Automation, triggerName: string) => {
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
    throw new Error('Failed to trigger automation');
  }
  return await res.json();
};
