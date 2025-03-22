export function toInt(str: string): number | null {
    const parsed = parseInt(str, 10); 
  
    if (isNaN(parsed)) {
      return null; 
    }
  
    return parsed;
  }