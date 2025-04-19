/**
 * Test file for Layer 2 (JavaScript Reactive Layer) MCP integration
 * 
 * This file contains comprehensive tests for the JavaScript MCP client,
 * verifying proper WebSocket communication, context conversion, error handling,
 * and reconnection capabilities.
 */

const assert = require('assert');
const WebSocket = require('ws');
const EventEmitter = require('events');

// Mock the actual client implementation for testing
const MCPClient = require('../src/main/javascript/mcp_client_layer2.js');
const NodeBridge = require('../../core/reactive/node_bridge');

// Mock server for testing WebSocket communication
class MockMCPServer {
  constructor(port = 9000) {
    this.port = port;
    this.clients = new Set();
    this.messageLog = [];
    this.server = null;
    this.isRunning = false;
  }

  start() {
    if (this.isRunning) return;

    this.server = new WebSocket.Server({ port: this.port });
    
    this.server.on('connection', (socket) => {
      this.clients.add(socket);
      
      socket.on('message', (message) => {
        const parsed = JSON.parse(message);
        this.messageLog.push(parsed);
        
        // Echo back messages with server timestamp added
        const response = {
          ...parsed,
          serverTimestamp: Date.now(),
          type: parsed.type === 'REQUEST' ? 'RESPONSE' : parsed.type
        };
        
        socket.send(JSON.stringify(response));
      });
      
      socket.on('close', () => {
        this.clients.delete(socket);
      });
      
      // Send welcome message
      socket.send(JSON.stringify({
        type: 'SERVER_INFO',
        serverTime: Date.now(),
        version: '1.0.0',
        supportedLayers: [0, 1, 2, 3, 4, 5, 6]
      }));
    });
    
    this.isRunning = true;
    console.log(`Mock MCP Server running on port ${this.port}`);
    
    return this;
  }
  
  stop() {
    if (!this.isRunning) return;
    
    this.server.close();
    this.isRunning = false;
    this.clients.clear();
    this.messageLog = [];
    
    console.log('Mock MCP Server stopped');
  }
  
  // Simulate server disconnection for testing reconnection logic
  simulateDisconnect() {
    this.clients.forEach(client => {
      client._socket.destroy();
    });
    this.clients.clear();
  }
  
  getMessageLog() {
    return this.messageLog;
  }
  
  clearMessageLog() {
    this.messageLog = [];
  }
}

// Helper function to wait for a specified time
function sleep(ms) {
  return new Promise(resolve => setTimeout(resolve, ms));
}

// Test suite for MCP Layer 2 integration
describe('Layer 2 MCP Integration Tests', () => {
  let mockServer;
  let client;
  
  // Setup mock server before running tests
  before(() => {
    mockServer = new MockMCPServer(9001).start();
  });
  
  // Cleanup after tests
  after(() => {
    mockServer.stop();
  });
  
  // Reset before each test
  beforeEach(() => {
    mockServer.clearMessageLog();
    client = new MCPClient({
      serverUrl: 'ws://localhost:9001',
      reconnectInterval: 1000,
      heartbeatInterval: 5000
    });
  });
  
  // Clean up after each test
  afterEach(() => {
    if (client) {
      client.disconnect();
      client = null;
    }
  });
  
  // Test basic connection establishment
  it('should establish connection with MCP server', async () => {
    await client.connect();
    await sleep(500); // Give time for connection to establish
    
    assert.strictEqual(client.isConnected(), true, 'Client should be connected');
    assert.strictEqual(client.connectionState, 'CONNECTED', 'Connection state should be CONNECTED');
    
    // Verify registration message was sent
    const messages = mockServer.getMessageLog();
    const registrationMessage = messages.find(msg => msg.type === 'REGISTER');
    
    assert.ok(registrationMessage, 'Registration message should be sent');
    assert.strictEqual(registrationMessage.layer, 2, 'Layer should be 2');
    assert.ok(Array.isArray(registrationMessage.capabilities), 'Capabilities should be an array');
  });
  
  // Test message sending with context
  it('should send messages with proper 7D context', async () => {
    await client.connect();
    await sleep(500);
    
    const context = {
      WHO: { id: 'user123', role: 'admin' },
      WHAT: { action: 'test', resource: 'system' },
      WHEN: { timestamp: Date.now() },
      WHERE: /Users/Jubicudis/TNOS1/Tranquility-Neuro-OS/mcp
      WHY: { purpose: 'testing' },
      HOW: { method: 'automated' },
      TO_WHAT_EXTENT: { scope: 'unit-test', priority: 'high' }
    };
    
    client.sendWithContext('TEST_MESSAGE', { data: 'test-payload' }, context);
    await sleep(500);
    
    const messages = mockServer.getMessageLog();
    const testMessage = messages.find(msg => msg.type === 'TEST_MESSAGE');
    
    assert.ok(testMessage, 'Test message should be sent');
    assert.deepStrictEqual(testMessage.context, context, 'Context should match exactly');
    assert.deepStrictEqual(testMessage.payload, { data: 'test-payload' }, 'Payload should match');
    assert.strictEqual(testMessage.source.layer, 2, 'Source layer should be 2');
  });
  
  // Test message receiving
  it('should receive and process messages from server', async () => {
    let receivedMessage = null;
    
    await client.connect();
    await sleep(500);
    
    // Set up message handler
    client.on('message', (message) => {
      receivedMessage = message;
    });
    
    // Trigger a message from the server (via echo)
    client.send('ECHO', { echo: 'test-echo' });
    await sleep(500);
    
    assert.ok(receivedMessage, 'Message should be received');
    assert.strictEqual(receivedMessage.payload.echo, 'test-echo', 'Echo payload should match');
    assert.ok(receivedMessage.serverTimestamp, 'Server timestamp should be present');
  });
  
  // Test automatic reconnection
  it('should automatically reconnect after connection loss', async () => {
    await client.connect();
    await sleep(500);
    
    assert.strictEqual(client.isConnected(), true, 'Client should be connected initially');
    
    // Simulate server disconnect
    mockServer.simulateDisconnect();
    await sleep(100);
    
    assert.strictEqual(client.isConnected(), false, 'Client should be disconnected after server disconnect');
    
    // Restart mock server
    mockServer.start();
    
    // Wait for reconnection
    await sleep(2000);
    
    assert.strictEqual(client.isConnected(), true, 'Client should reconnect automatically');
    assert.ok(client.reconnectionAttempts > 0, 'Reconnection attempts should be tracked');
  });
  
  // Test message queueing when disconnected
  it('should queue messages when disconnected and send upon reconnection', async () => {
    await client.connect();
    await sleep(500);
    
    // Disconnect client
    mockServer.simulateDisconnect();
    await sleep(100);
    
    // Send messages while disconnected
    client.send('OFFLINE_MESSAGE', { id: 1 });
    client.send('OFFLINE_MESSAGE', { id: 2 });
    client.send('OFFLINE_MESSAGE', { id: 3 });
    
    // Verify messages are queued
    assert.strictEqual(client.messageQueue.length, 3, 'Messages should be queued when offline');
    
    // Restart mock server and allow reconnection
    mockServer.start();
    await sleep(2000);
    
    // Verify queued messages were sent
    const messages = mockServer.getMessageLog();
    const offlineMessages = messages.filter(msg => msg.type === 'OFFLINE_MESSAGE');
    
    assert.strictEqual(offlineMessages.length, 3, 'All queued messages should be sent after reconnection');
    assert.strictEqual(client.messageQueue.length, 0, 'Message queue should be empty after sending');
  });
  
  // Test heartbeat mechanism
  it('should send heartbeat messages at regular intervals', async () => {
    // Configure shorter heartbeat interval for testing
    client = new MCPClient({
      serverUrl: 'ws://localhost:9001',
      reconnectInterval: 1000,
      heartbeatInterval: 1000 // 1 second heartbeat for faster testing
    });
    
    await client.connect();
    
    // Wait for multiple heartbeat intervals
    await sleep(3500);
    
    const messages = mockServer.getMessageLog();
    const heartbeats = messages.filter(msg => msg.type === 'HEARTBEAT');
    
    assert.ok(heartbeats.length >= 3, 'Multiple heartbeat messages should be sent');
    
    // Verify increasing timestamps in heartbeats
    for (let i = 1; i < heartbeats.length; i++) {
      assert.ok(
        heartbeats[i].timestamp > heartbeats[i-1].timestamp,
        'Heartbeat timestamps should be increasing'
      );
    }
  });
  
  // Test 7D context conversion
  it('should properly convert JavaScript objects to 7D context', () => {
    const jsObject = {
      user: { id: 'user123', role: 'admin' },
      action: 'click',
      timestamp: Date.now(),
      element: '#submit-button',
      purpose: 'form-submission',
      method: 'ajax',
      priority: 'high'
    };
    
    const context = client.convertToContext(jsObject);
    
    assert.ok(context.WHO, 'WHO dimension should be present');
    assert.ok(context.WHAT, 'WHAT dimension should be present');
    assert.ok(context.WHEN, 'WHEN dimension should be present');
    assert.ok(context.WHERE/Users/Jubicudis/TNOS1/Tranquility-Neuro-OS/mcp
    assert.ok(context.WHY, 'WHY dimension should be present');
    assert.ok(context.HOW, 'HOW dimension should be present');
    assert.ok(context.TO_WHAT_EXTENT, 'TO_WHAT_EXTENT dimension should be present');
    
    assert.strictEqual(context.WHO.id, 'user123', 'User ID should be mapped to WHO dimension');
    assert.strictEqual(context.WHAT.action, 'click', 'Action should be mapped to WHAT dimension');
    assert.strictEqual(context.WHERE/Users/Jubicudis/TNOS1/Tranquility-Neuro-OS/mcp
    assert.strictEqual(context.WHY.purpose, 'form-submission', 'Purpose should be mapped to WHY dimension');
    assert.strictEqual(context.HOW.method, 'ajax', 'Method should be mapped to HOW dimension');
    assert.strictEqual(context.TO_WHAT_EXTENT.priority, 'high', 'Priority should be mapped to TO_WHAT_EXTENT dimension');
  });
  
  // Test error handling
  it('should handle errors appropriately', async () => {
    let errorHandled = false;
    let errorType = null;
    
    // Override error handler for testing
    client._handleError = (error, context, retry) => {
      errorHandled = true;
      errorType = error.type;
      // Don't retry in test
    };
    
    await client.connect();
    await sleep(500);
    
    // Simulate a validation error
    client._processMessage({
      type: 'ERROR',
      error: {
        type: 'VALIDATION',
        message: 'Invalid message format'
      }
    });
    
    assert.strictEqual(errorHandled, true, 'Error should be handled');
    assert.strictEqual(errorType, 'VALIDATION', 'Error type should be correctly identified');
  });
  
  // Test Node.js Bridge functionality
  it('should provide Node.js Bridge functionality', async () => {
    const bridge = new NodeBridge({
      serverUrl: 'ws://localhost:9001',
      reconnectInterval: 1000
    });
    
    let receivedEvent = null;
    
    // Set up event handler
    bridge.on('message', (message) => {
      receivedEvent = message;
    });
    
    await sleep(1000); // Wait for bridge connection
    
    // Send a message through the bridge
    bridge.send('BRIDGE_TEST', { source: 'node-app' });
    await sleep(500);
    
    // Verify message was sent
    const messages = mockServer.getMessageLog();
    const bridgeMessage = messages.find(msg => msg.type === 'BRIDGE_TEST');
    
    assert.ok(bridgeMessage, 'Message should be sent through bridge');
    assert.strictEqual(bridgeMessage.payload.source, 'node-app', 'Payload should be correct');
    
    // Clean up
    bridge.client.disconnect();
  });
  
  // Test batch processing
  it('should support message batching for efficiency', async () => {
    client = new MCPClient({
      serverUrl: 'ws://localhost:9001',
      batchMessages: true,
      batchSize: 3,
      batchInterval: 500
    });
    
    await client.connect();
    await sleep(500);
    
    // Send multiple messages in quick succession
    client.send('BATCH_TEST', { id: 1 });
    client.send('BATCH_TEST', { id: 2 });
    client.send('BATCH_TEST', { id: 3 });
    
    await sleep(1000); // Wait for batch to be processed
    
    const messages = mockServer.getMessageLog();
    const batchMessages = messages.filter(msg => msg.type === 'BATCH');
    
    assert.ok(batchMessages.length > 0, 'Batch message should be sent');
    const batch = batchMessages[0];
    
    assert.ok(Array.isArray(batch.messages), 'Batch should contain messages array');
    assert.strictEqual(batch.messages.length, 3, 'Batch should contain all messages');
    
    // Verify batch contents
    const ids = batch.messages
      .filter(msg => msg.type === 'BATCH_TEST')
      .map(msg => msg.payload.id);
    
    assert.deepStrictEqual(ids.sort(), [1, 2, 3], 'All messages should be in the batch');
  });
});

// Export for automated test runners
module.exports = {
  MockMCPServer,
  runTests: () => {
    describe('Run all Layer 2 MCP integration tests', () => {
      // Tests are defined above
    });
  }
};

