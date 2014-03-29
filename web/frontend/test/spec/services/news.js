'use strict';

describe('Service: News', function () {

  // load the service's module
  beforeEach(module('newshoundApp'));

  // instantiate service
  var News;
  beforeEach(inject(function (_News_) {
    News = _News_;
  }));

  it('should do something', function () {
    expect(!!News).toBe(true);
  });

});
