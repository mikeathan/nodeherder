import { getErrorMessage } from '@/contracts/errors';
import { ApiResponse } from '@/types/api.type';
import { Automation } from '@/types/automation.type';
const baseUrl = import.meta.env.VITE_API_BASE_URL;

export const triggerAutomation = async (automation: Automation, triggerName: string): Promise<ApiResponse<string>> => {
  try {
    const automationId = automation.id;
    console.info(`Triggering automation: ${automationId} with target: ${triggerName}`);

    const res = await fetch(`${baseUrl}/api/automation/trigger`, {
      method: 'POST',
      headers: {
        'Content-Type': 'application/json',
      },
      body: JSON.stringify({ triggerName, automationId }),
    });

    if (!res.ok) {
      console.error('Response failed to trigger automation', res);
      const errorData = await res.json();
      return { success: false, error: errorData.error };
    }
  } catch (error: unknown) {
    console.error('Error: Failed to trigger automation', error);
    return { success: false, error: getErrorMessage(error) };
  }

  return { success: true };
};
