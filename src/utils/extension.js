/**
 * WHO: TNOSVSCodeExtension
 * WHAT: VS Code Extension for TNOS Integration
 * WHEN: VS Code startup and runtime
 * WHERE: /Users/Jubicudis/TNOS1/Tranquility-Neuro-OS/mcp
 * WHY: To provide TNOS capabilities through VS Code
 * HOW: Using Language Server Protocol with custom adapter
 * EXTENT: All VS Code <-> TNOS communication
 */

const vscode = require('vscode');
const path = require('path');
const { workspace, ExtensionContext, window, commands } = vscode;
const { LanguageClient, TransportKind } = require('../src/main/javascript/node.js');

let client;
let outputChannel;

/**
 * WHO: ExtensionActivator
 * WHAT: Extension activation with protocol adapter
 * WHEN: VS Code startup
 * WHERE: /Users/Jubicudis/TNOS1/Tranquility-Neuro-OS/mcp
 * WHY: To establish TNOS MCP connection
 * HOW: Using custom protocol adapter
 * EXTENT: Extension lifecycle
 */
function activate(context) {
    // Create output channel for TNOS messages
    outputChannel = window.createOutputChannel('TNOS MCP');
    outputChannel.show();
    outputChannel.appendLine('TNOS MCP Extension activated');
    
    // Log with 7D context
    log7D('info', 'Extension activated', {
        who: 'TNOSVSCodeExtension',
        what: 'Activation',
        when: new Date().toISOString(),
        where: /Users/Jubicudis/TNOS1/Tranquility-Neuro-OS/mcp
        why: 'StartupSequence',
        how: 'VSCodeAPI',
        extent: 'ExtensionLifecycle'
    });

    try {
        // Get adapter path
        const adapterPath = path.join(context.extensionPath, 'vscode_protocol_adapter.js');
        
        // Server options - using the protocol adapter
        const serverOptions = {
            run: {
                command: adapterPath,
                transport: TransportKind.stdio,
                args: []
            },
            debug: {
                command: adapterPath,
                transport: TransportKind.stdio,
                args: ['--debug']
            }
        };

        // Client options
        const clientOptions = {
            documentSelector: [
                { scheme: 'file', language: 'javascript' },
                { scheme: 'file', language: 'typescript' },
                { scheme: 'file', language: 'python' },
                { scheme: 'file', language: 'java' },
                { scheme: 'file', language: 'cpp' }
            ],
            synchronize: {
                fileEvents: workspace.createFileSystemWatcher('**/*')
            },
            outputChannel: outputChannel,
            initializationOptions: {
                tnos: {
                    mcpBridgePort: 8080,
                    mcpServerPort: 8888,
                    contextEnabled: true,
                    compressionEnabled: true
                }
            }
        };

        // Create the language client
        client = new LanguageClient(
            'tnos-mcp',
            'TNOS MCP',
            serverOptions,
            clientOptions
        );

        // Register commands
        registerCommands(context);

        // Start the client
        client.start();
        log7D('info', 'TNOS MCP client started', {
            what: 'ClientStartup',
            why: 'EstablishConnection',
            how: 'LSPClient'
        });
    } catch (error) {
        log7D('error', `Failed to start TNOS MCP client: ${error.message}`, {
            what: 'ClientStartupFailure',
            why: 'ErrorHandling',
            how: 'ExceptionProcessing',
            extent: 'StartupPhase'
        });
        window.showErrorMessage(`TNOS MCP: Failed to start client: ${error.message}`);
    }
}

/**
 * WHO: CommandRegistrator
 * WHAT: Register VS Code commands for TNOS integration
 * WHEN: During extension activation
 * WHERE: /Users/Jubicudis/TNOS1/Tranquility-Neuro-OS/mcp
 * WHY: To provide TNOS functionality through commands
 * HOW: Using VS Code command registration
 * EXTENT: All TNOS commands
 */
function registerCommands(context) {
    // TNOS health check command
    context.subscriptions.push(commands.registerCommand('tnos.healthCheck', async () => {
        try {
            const result = await client.sendRequest('tnos/healthCheck');
            outputChannel.appendLine('TNOS Health Check:');
            outputChannel.appendLine(JSON.stringify(result, null, 2));
            window.showInformationMessage('TNOS is healthy and connected');
            return result;
        } catch (error) {
            window.showErrorMessage(`TNOS Health Check failed: ${error.message}`);
            return { status: 'error', message: error.message };
        }
    }));

    // TNOS compression command
    context.subscriptions.push(commands.registerCommand('tnos.compress', async () => {
        const editor = window.activeTextEditor;
        if (!editor) {
            window.showInformationMessage('No editor is active');
            return;
        }

        const text = editor.document.getText(editor.selection);
        if (!text) {
            window.showInformationMessage('No text selected');
            return;
        }

        try {
            const result = await client.sendRequest('tnos/compress', { content: text });
            outputChannel.appendLine('Compression result:');
            outputChannel.appendLine(JSON.stringify(result, null, 2));
            window.showInformationMessage(`Compressed with ratio: ${result.ratio}`);
            return result;
        } catch (error) {
            window.showErrorMessage(`TNOS Compression failed: ${error.message}`);
            return { status: 'error', message: error.message };
        }
    }));

    // TNOS context analysis command
    context.subscriptions.push(commands.registerCommand('tnos.analyzeContext', async () => {
        const editor = window.activeTextEditor;
        if (!editor) {
            window.showInformationMessage('No editor is active');
            return;
        }

        const text = editor.document.getText(editor.selection);
        if (!text) {
            window.showInformationMessage('No text selected');
            return;
        }

        try {
            const result = await client.sendRequest('tnos/analyzeContext', { content: text });
            outputChannel.appendLine('Context analysis result:');
            outputChannel.appendLine(JSON.stringify(result, null, 2));
            
            // Show a formatted view of the 7D context
            const panel = window.createWebviewPanel(
                'tnosContext',
                'TNOS Context Analysis',
                vscode.ViewColumn.Beside,
                { enableScripts: true }
            );
            
            panel.webview.html = createContextViewHtml(result);
            return result;
        } catch (error) {
            window.showErrorMessage(`TNOS Context Analysis failed: ${error.message}`);
            return { status: 'error', message: error.message };
        }
    }));
}

/**
 * WHO: ContextViewGenerator
 * WHAT: Generate HTML view for 7D context
 * WHEN: After context analysis
 * WHERE: /Users/Jubicudis/TNOS1/Tranquility-Neuro-OS/mcp
 * WHY: To visualize 7D context
 * HOW: Using HTML with embedded styles
 * EXTENT: Single context analysis result
 */
function createContextViewHtml(contextResult) {
    return `
        <!DOCTYPE html>
        <html lang="en">
        <head>
            <meta charset="UTF-8">
            <meta name="viewport" content="width=device-width, initial-scale=1.0">
            <title>TNOS Context Analysis</title>
            <style>
                body { font-family: Arial, sans-serif; padding: 20px; }
                .context-dimension { margin-bottom: 15px; }
                .dimension-name { font-weight: bold; color: #0078D4; }
                .dimension-value { margin-left: 10px; }
                h1 { color: #0078D4; }
                .context-container { 
                    border: 1px solid #ccc; 
                    border-radius: 5px; 
                    padding: 15px;
                    background-color: #f9f9f9;
                }
                .metadata { 
                    margin-top: 20px;
                    font-size: 0.9em;
                    color: #666;
                }
            </style>
        </head>
        <body>
            <h1>TNOS 7D Context Analysis</h1>
            <div class="context-container">
                <div class="context-dimension">
                    <span class="dimension-name">WHO:</span>
                    <span class="dimension-value">${contextResult.who || 'Not identified'}</span>
                </div>
                <div class="context-dimension">
                    <span class="dimension-name">WHAT:</span>
                    <span class="dimension-value">${contextResult.what || 'Not identified'}</span>
                </div>
                <div class="context-dimension">
                    <span class="dimension-name">WHEN:</span>
                    <span class="dimension-value">${contextResult.when || 'Not identified'}</span>
                </div>
                <div class="context-dimension">
                    <span class="dimension-name">WHERE:/Users/Jubicudis/TNOS1/Tranquility-Neuro-OS/mcp
                    <span class="dimension-value">${contextResult.where /Users/Jubicudis/TNOS1/Tranquility-Neuro-OS/mcp
                </div>
                <div class="context-dimension">
                    <span class="dimension-name">WHY:</span>
                    <span class="dimension-value">${contextResult.why || 'Not identified'}</span>
                </div>
                <div class="context-dimension">
                    <span class="dimension-name">HOW:</span>
                    <span class="dimension-value">${contextResult.how || 'Not identified'}</span>
                </div>
                <div class="context-dimension">
                    <span class="dimension-name">TO WHAT EXTENT:</span>
                    <span class="dimension-value">${contextResult.extent || 'Not identified'}</span>
                </div>
            </div>
            <div class="metadata">
                <p>Analysis generated: ${new Date().toISOString()}</p>
                <p>TNOS MCP Version: ${contextResult.metadata?.version || 'Unknown'}</p>
            </div>
        </body>
        </html>
    `;
}

/**
 * WHO: Log7DFunction
 * WHAT: Log messages with 7D context
 * WHEN: During extension operation
 * WHERE: /Users/Jubicudis/TNOS1/Tranquility-Neuro-OS/mcp
 * WHY: To maintain context-aware logging
 * HOW: Using output channel with context enrichment
 * EXTENT: All extension logging
 */
function log7D(level, message, context = {}) {
    // Default context
    const defaultContext = {
        who: 'TNOSVSCodeExtension',
        what: 'LogOperation',
        when: new Date().toISOString(),
        where: /Users/Jubicudis/TNOS1/Tranquility-Neuro-OS/mcp
        why: 'ContextualLogging',
        how: 'OutputChannel',
        extent: 'SingleLogEntry'
    };

    // Merge with provided context
    const fullContext = { ...defaultContext, ...context };
    
    // Format log message with context
    const formattedMessage = `[${level.toUpperCase()}] [${fullContext.when}] [${fullContext.who}] ${message}`;
    
    // Log to output channel
    if (outputChannel) {
        outputChannel.appendLine(formattedMessage);
    }
}

/**
 * WHO: ExtensionDeactivator
 * WHAT: Clean up and shut down extension
 * WHEN: VS Code shutdown or extension deactivation
 * WHERE: /Users/Jubicudis/TNOS1/Tranquility-Neuro-OS/mcp
 * WHY: To properly close resources
 * HOW: Using client stop method
 * EXTENT: Extension cleanup phase
 */
function deactivate() {
    log7D('info', 'Extension deactivating', {
        what: 'Deactivation',
        why: 'ShutdownSequence'
    });
    
    if (client) {
        return client.stop();
    }
    return undefined;
}

module.exports = {
    activate,
    deactivate
};


