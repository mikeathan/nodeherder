import { Automation } from '@/types/automation.type';
import axios from 'axios';

const triggerAutomation = async (automation: Automation, triggerName: string) => {
  var triggerId = triggerName;
  var automationId = automation.id;
  const response = await axios.post('/api/automation/trigger', { triggerId, automationId });
  return response.data;
};
