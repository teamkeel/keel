import { createMockPool, createMockQueryResult } from 'slonik';
import Query from './query';

interface Post {
  title: string
  subTitle: string
}

describe('QueryAPI', () => {
  describe('execution', () => {
    describe('multiple wheres', () => {
      it('returns the expected results', async () => {
        const tableName = 'posts';

        const pool = createMockPool({
          query: async() => {
            return createMockQueryResult([
              {
                title: 'Foo',
                subTitle: 'Bar'
              }
            ]);
          }
        });

        const queryBuilder = new Query<Post>({ tableName, pool });
      
        const query = queryBuilder.where({
          title: {
            endsWith: 'hello'
          },
          subTitle: {
            contains: 'hehe'
          }
        }).orWhere({
          title: 'djujd'
        });
  
        expect(await query.all())
          .toEqual([
            {
              title: 'Foo',
              subTitle: 'Bar'
            }
          ]);
      });
    });
  });
});
