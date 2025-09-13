#!/usr/bin/env node

const WebSocket = require('ws');

// Test WebSocket connection to KubeChat API
const WS_URL = 'ws://localhost:30080/api/v1/ws';

console.log('🔗 Testing WebSocket connection to KubeChat API...');
console.log(`📡 Connecting to: ${WS_URL}`);

const ws = new WebSocket(WS_URL, {
    headers: {
        'User-Agent': 'KubeChat-Test-Client/1.0'
    }
});

ws.on('open', function open() {
    console.log('✅ WebSocket connection established!');
    
    // Test authentication
    console.log('🔐 Testing authentication...');
    const authMessage = {
        type: 'auth',
        payload: {
            token: 'test-invalid-token'  // This should fail but test the flow
        }
    };
    
    ws.send(JSON.stringify(authMessage));
    
    // Set timeout to close connection
    setTimeout(() => {
        console.log('🔚 Closing connection...');
        ws.close();
    }, 3000);
});

ws.on('message', function message(data) {
    console.log('📨 Received message:', data.toString());
    
    try {
        const msg = JSON.parse(data.toString());
        console.log('📋 Parsed message:', {
            type: msg.type,
            payload: msg.payload ? Object.keys(msg.payload) : 'none'
        });
    } catch (e) {
        console.log('⚠️  Failed to parse message as JSON');
    }
});

ws.on('error', function error(err) {
    console.error('❌ WebSocket error:', err.message);
});

ws.on('close', function close(code, reason) {
    console.log(`🔚 Connection closed with code: ${code}, reason: ${reason}`);
    console.log('✅ WebSocket test completed');
});

// Handle process termination
process.on('SIGINT', () => {
    console.log('\n🛑 Test interrupted by user');
    ws.close();
    process.exit(0);
});