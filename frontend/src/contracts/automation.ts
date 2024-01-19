export type Automation = {
    id: string,
    friendlyname: string,
    description: string,
    enabled: boolean,
    triggers: Array<AutomationTrigger>
}
export type AutomationTrigger = {
    name: string,
    conditions: Array<AutomationTriggerCondition>,
    action: AutomationTriggerAction
}

export type AutomationTriggerCondition = {
    name: string,
    value: any | null,
    equality: string,
}

export type AutomationTriggerAction = {
    id: string,
    friendlyname: string,
    property: string,
    data: any | undefined,
    operation: number | undefined,
}
