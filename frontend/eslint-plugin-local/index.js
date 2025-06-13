export default {
  rules: {
    'no-nested-interactive': {
      meta: {
        type: 'problem',
        docs: {
          description: 'Disallow nested interactive elements (buttons inside trigger components)',
          category: 'Best Practices',
          recommended: true
        },
        messages: {
          nestedInteractive:
            'Avoid nesting {{inner}} inside {{outer}}. {{outer}} already renders an interactive element. Use the child snippet pattern or pass content directly.\n\nExample fix:\nBefore:\n<DialogTrigger>\n  <Button variant="outline">\n    <Icon />\n    <span>Text</span>\n  </Button>\n</DialogTrigger>\n\nAfter:\n<DialogTrigger class={buttonVariants({ variant: "outline" })}>\n  <Icon />\n  <span>Text</span>\n</DialogTrigger>'
        },
        schema: []
      },
      create(context) {
        const componentStack = [];

        // List of trigger components that render as buttons
        // Includes both standalone (DialogTrigger) and dot notation (Dialog.Trigger) names
        const triggerComponents = [
          'DialogTrigger',
          'Trigger', // Catches all .Trigger variants
          'AlertDialogTrigger',
          'PopoverTrigger',
          'DropdownMenuTrigger',
          'AccordionTrigger',
          'CollapsibleTrigger',
          'SelectTrigger',
          'NavigationMenuTrigger',
          'TabsTrigger',
          'MenubarTrigger',
          'SheetTrigger',
          'DrawerTrigger',
          'HoverCardTrigger',
          'ContextMenuTrigger',
          'TooltipTrigger'
        ];

        // List of components that are interactive elements
        const interactiveComponents = [
          'Button',
          'SidebarTrigger',
          'FormButton',
          'SidebarMenuButton',
          'Checkbox',
          'Switch',
          'Toggle',
          'RadioGroupItem'
        ];

        function getComponentName(node) {
          if (node.name) {
            // Handle simple component names like Button, DialogTrigger
            if (node.name.type === 'Identifier') {
              return node.name.name;
            }
            // Handle dot notation like Dialog.Trigger, Popover.Trigger
            if (node.name.type === 'SvelteMemberExpressionName' && node.name.property) {
              return node.name.property.name;
            }
          }
          return null;
        }

        return {
          'SvelteElement[kind="component"]'(node) {
            const componentName = getComponentName(node);

            if (componentName) {
              // Track trigger components
              if (triggerComponents.includes(componentName)) {
                componentStack.push(componentName);
              }

              // Check for interactive components inside triggers
              if (interactiveComponents.includes(componentName)) {
                const parentTrigger = componentStack.find((comp) =>
                  triggerComponents.includes(comp)
                );
                if (parentTrigger) {
                  context.report({
                    node,
                    messageId: 'nestedInteractive',
                    data: {
                      inner: componentName,
                      outer: parentTrigger
                    }
                  });
                }
              }
            }
          },
          'SvelteElement[kind="component"]:exit'(node) {
            const componentName = getComponentName(node);

            if (componentName) {
              // Pop trigger component from stack when exiting
              if (triggerComponents.includes(componentName)) {
                const index = componentStack.lastIndexOf(componentName);
                if (index !== -1) {
                  componentStack.splice(index, 1);
                }
              }
            }
          }
        };
      }
    }
  }
};
