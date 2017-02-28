Hi, @christian-bromann

I will try to explain why these changes are needed.


### Now webdriverio architecture (simplistically) consists of two components:

- one Launcher
- some Runners

1. The Launcher initialized reporters and fork some runners. Then it subscribes to the event 'message':

   ```javascript
   // launcher.js
   childProcess
               .on('message', this.messageHandler.bind(this, cid))
               .on('exit', this.endHandler.bind(this, rid))
   ```

2. In the processes, we forked, we send messages using `process.send()` (see the `runner.js` file).
3. Launcher process receives them, parses, and sends them to reporters. In the reporters, we are processing these messages and generate reports.

Each message that we send from child processes (in the `runner.js` file), contains a field `cid`. We determine the value of this field only once when the runner process starts and it no longer changes. So I want to bind this value to each child (runner) process in the parent (launcher) process.

This will allow us to send custom messages from the child (runner) processes and handle them into reporters.

It is now possible to write custom reporters. To do this, we need to register an event handler on several events which get triggered during the test. With my changes we will be able to create custom test-events and handle it in the custom reporters.

### For example

__We have a test:__

```javascript
describe('Our awesome test', () => {
    it('test super feature', () => {
        process.send({
            event: 'add-feature-to-report',
            feature: 'super-feature'
        });
        
        assert.equals([] + {}, '[object Object]', 'Wat?');
    })
});
```

__And in reporter we register event handler:__

```javascript
this.on('add-feature-to-report', ({event, cid, feature}) => {
    console.log(`Process ${cid}. We tests feature: ${feature}`);
});
```

__Output with my changes:__

```
Process 0-0. We tests feature: super-feature
```

__Without my changes:__

```
Process undefined. We tests feature: super-feature
```

Since the the driver can run multiple tests in parallel, the second case is bad. We can not determine the mapping between features done and tests.

### Why do I need it?

I am developing a reporter for [Allure Framework](https://github.com/allure-framework/allure-mocha). And I need the ability to create your own test-events, such as in the example above.

